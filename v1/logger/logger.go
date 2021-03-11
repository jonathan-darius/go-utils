package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"net"
	"strings"

	"github.com/forkyid/go-utils/v1/uuid"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v3"
)

// Init initialize logger
func Init() {
	if os.Getenv("ENV") == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}
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

func log(service string, fields logrus.Fields, errMsg string) {
	logger := logrus.New()

	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(os.Getenv("ELASTICSEARCH_HOST")),
	)
	if err != nil {
		logger.Println("logger: ", err.Error())
		logger.Println(errMsg)
		return
	}

	hook, err := elogrus.NewAsyncElasticHook(
		client,
		os.Getenv("ELASTICSEARCH_HOST"),
		logrus.DebugLevel,
		service,
	)
	if err != nil {
		logger.Println("logger: ", err.Error())
		logger.Println(errMsg)
		return
	}
	logger.Hooks.Add(hook)

	start := time.Now()
	latency := time.Since(start)
	fields["ResponseTime"] = latency

	// get the callers (depth 2)
	stack := ""
	indent := ""
	for i := 3; i > 1; i-- {
		pc, file, line, ok := runtime.Caller(i)
		if ok {
			stack += fmt.Sprintf("%s%s %s#%d\n", indent, runtime.FuncForPC(pc).Name(), file, line)
		}
		indent += "\t"
	}
	fields["Trace"] = stack

	logger.WithFields(fields).Error(errMsg)
}

// LogWithContext for api
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

	log("service-logger", fields, errMsg)
}

// Log for consumer
func Log(errMsg string) {
	fields := logrus.Fields{
		"Key":         uuid.GetUUID(),
		"ServiceName": os.Getenv("SERVICE_NAME"),
	}

	log("service-consumer-logger", fields, errMsg)
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
