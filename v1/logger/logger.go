package logger

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/forkyid/go-utils/v1/uuid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var cidrs []*net.IPNet

const (
	envProduction = "production"
	tagJson       = "json"
	tagLogIgnore  = "logignore"
)

func init() {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		if os.Getenv("ENV") == envProduction {
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetOutput(os.Stdout)
		} else {
			logger.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
				ForceColors:   true,
				DisableQuote:  true,
			})
		}
	}
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}
	cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}
}

func isPrivateAddress(address string) (isPrivate bool, err error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		isPrivate = false
		err = errors.New("address is not valid")
		return
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			isPrivate = true
			return
		}
	}

	isPrivate = false
	return
}

func realIP(req *http.Request) (ip string) {
	ip = req.Header.Get("X-Real-Ip")
	xForwardedFor := req.Header.Get("X-Forwarded-For")
	if ip == "" && xForwardedFor == "" {
		if strings.ContainsRune(req.RemoteAddr, ':') {
			ip, _, _ = net.SplitHostPort(req.RemoteAddr)
		} else {
			ip = req.RemoteAddr
		}
		return
	}

	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		isPrivate, err := isPrivateAddress(address)
		if !isPrivate && err == nil {
			ip = address
			return
		}
	}
	return
}

func traceStack() (stack string) {
	stack = ""
	ut, _ := template.New("stack").Parse("\n\t{{ .Name }} {{ .File }}:{{ .Line }}")
	for i := 1; i < 4; i++ {
		if pc, file, line, ok := runtime.Caller(i); ok {
			buf := new(bytes.Buffer)
			ut.Execute(buf, struct {
				Name string
				File string
				Line int
			}{
				Name: runtime.FuncForPC(pc).Name(),
				File: file,
				Line: line,
			})
			stack += buf.String()
		}
	}
	return
}

func parseBody(v interface{}) (body map[string]string) {
	fmt.Println(v, reflect.TypeOf(v).Kind(), reflect.ValueOf(v).Len())
	if v == nil || reflect.TypeOf(v).Kind() != reflect.Slice || reflect.ValueOf(v).Len() == 0 {
		fmt.Println("return first")
		return
	}

	val := reflect.ValueOf(v).Index(0)
	fmt.Println(val.Kind(), val.Elem().Kind())
	if val.Kind() == reflect.Interface && val.Elem().Kind() == reflect.Struct {
		val = val.Elem()
	} else {
		fmt.Println("return second")
		return
	}

	t := val.Type()
	body = map[string]string{}
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		logIgnore := field.Tag.Get(tagLogIgnore)
		if logIgnore == "true" {
			continue
		}
		body[field.Tag.Get(tagJson)] = val.Field(i).String()
	}
	return
}

func defineFields(ctx *gin.Context, args ...interface{}) (fields logrus.Fields) {
	if ctx == nil {
		return
	}

	params := map[string]string{}
	for _, p := range ctx.Params {
		params[p.Key] = p.Value
	}
	fields = logrus.Fields{
		"Key":         uuid.GetUUID(),
		"ServiceName": os.Getenv("SERVICE_NAME"),
		"Params":      params,
		"StatusCode":  ctx.Writer.Status(),
		"Trace":       traceStack(),
	}
	req := ctx.Request
	if req != nil {
		fields["Request"] = req.RequestURI
		fields["Method"] = req.Method
		fields["IP"] = realIP(req)
		fields["RemoteAddress"] = req.Header.Get("X-Request-Id")
	}
	if len(args) > 0 {
		fields["Body"] = parseBody(args[0])
	}
	return
}

// Fatalf params
//	@ctx: *gin.Context
//	@errMsg: string
//	@err: error
//	@args: ...interface{}
func Fatalf(ctx *gin.Context, errMsg string, err error, args ...interface{}) {
	logger.WithFields(defineFields(ctx, args)).Fatalf(errMsg+": %v", err)
}

// Errorf params
//	@ctx: *gin.Context
//	@errMsg: string
//	@err: error
//	@args: ...interface{}
func Errorf(ctx *gin.Context, errMsg string, err error, args ...interface{}) {
	logger.WithFields(defineFields(ctx, args)).Errorf(errMsg+": %v", err)
}

// Warnf params
//	@errMsg: string
//	@err: error
func Warnf(errMsg string, err error) {
	logger.WithFields(logrus.Fields{"Trace": traceStack()}).Warnf(errMsg+": %v", err)
}

// Infof params
//	@errMsg: string
func Infof(errMsg string) {
	logger.WithFields(logrus.Fields{}).Info(errMsg)
}

// Debugf params
//	@ctx: *gin.Context
//	@errMsg: string
//	@err: error
//	@args: ...interface{}
func Debugf(ctx *gin.Context, errMsg string, err error, args ...interface{}) {
	logger.WithFields(defineFields(ctx, args)).Debugf(errMsg+": %v", err)
}
