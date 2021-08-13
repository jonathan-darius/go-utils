package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"text/template"

	"net"
	"strings"

	"github.com/forkyid/go-utils/v1/uuid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var l *logrus.Logger

// logger func
// 	return *logrus.Logger
func logger() *logrus.Logger {
	if l == nil {
		l = logrus.New()
		if os.Getenv("ENV") == "production" {
			l.SetFormatter(&logrus.JSONFormatter{})
			l.SetOutput(os.Stdout)
		} else {
			l.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
				ForceColors:   true,
				DisableQuote:  true,
			})
		}
	}

	return l
}

// realIP get the real IP from http request
func realIP(req *http.Request) string {
	ra := req.RemoteAddr
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// log params
// 	@fields: logrus.Fields
// 	@errMsg: string
func log(fields logrus.Fields, errMsg string) {
	logger := logger()

	stack := ""
	ut, _ := template.New("stack").Parse("\n\t{{ .Name }} {{ .File }}#{{ .Line }}")
	for i := 3; i > 1; i-- {
		pc, file, line, ok := runtime.Caller(i)
		if ok {
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
	fields["Trace"] = stack

	logger.WithFields(fields).Error(errMsg)
}

// LogWithContext params
// 	@c: *gin.Context
//	@uuid: string
// 	@errMsg: string
func LogWithContext(c *gin.Context, uuid, errMsg string) {
	req := c.Request
	payload := map[string]string{}
	for _, p := range c.Params {
		payload[p.Key] = p.Value
	}

	fields := logrus.Fields{
		"Key":         uuid,
		"ServiceName": os.Getenv("SERVICE_NAME"),
		"Payload":     payload,
		"StatusCode":  c.Writer.Status(),
	}

	if req != nil {
		fields["Request"] = req.RequestURI
		fields["Method"] = req.Method
		fields["IP"] = realIP(req)
		fields["RemoteAddress"] = req.Header.Get("X-Request-Id")
	}

	log(fields, errMsg)
}

// Log params
//	@errMsg: string
func Log(errMsg string) {
	fields := logrus.Fields{
		"Key":         uuid.GetUUID(),
		"ServiceName": os.Getenv("SERVICE_NAME"),
	}

	log(fields, errMsg)
}

// LogUserActivity for CMS user activity
func LogUserActivity(eventName, before, after, auth string) error {
	payload := map[string]string{
		"name":   eventName,
		"before": before,
		"after":  after,
	}

	payloadMarshal, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%v/cms/member/v1/activities", os.Getenv("API_ORIGIN_URL"))
	method := http.MethodPost
	headers := map[string]string{
		"Authorization": auth,
	}
	body := bytes.NewReader(payloadMarshal)

	req, _ := http.NewRequest(method, url, body)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error on inserting User Activities")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not OK")
	}

	return nil
}
