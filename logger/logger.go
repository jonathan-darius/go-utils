package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"net"
	"strings"

	"github.com/forkyid/go-utils/uuid"
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

type loggingResponseWriter struct {
	gin.ResponseWriter
	statusCode int
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

func newLoggingResponseWriter(w gin.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, w.Status()}
}

// LogError for api
func LogError(c *gin.Context, uuid, errMsg string) {
	w := c.Writer
	r := c.Request
	log := logrus.New()
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(os.Getenv("ELASTICSEARCH_HOST")))
	if err != nil {
		log.Println(err)
		log.Println(errMsg)
		return
	}

	hook, err := elogrus.NewAsyncElasticHook(client, os.Getenv("ELASTICSEARCH_HOST"), logrus.DebugLevel, "service-logger")
	if err != nil {
		log.Println(err)
		log.Println(errMsg)
		return
	}
	log.Hooks.Add(hook)

	start := time.Now()
	lw := newLoggingResponseWriter(w)

	latency := time.Since(start)

	payload := map[string]string{}
	for _, p := range c.Params {
		payload[p.Key] = p.Value
	}

	fields := logrus.Fields{
		"Key":          uuid,
		"ServiceName":  os.Getenv("SERVICE_NAME"),
		"Payload":      payload,
		"StatusCode":   lw.statusCode,
		"ResponseTime": latency,
	}

	if r != nil {
		fields["Request"] = r.RequestURI
		fields["Method"] = r.Method
		fields["IP"] = realIP(r)
		fields["RemoteAddress"] = r.Header.Get("X-Request-Id")
	}

	log.WithFields(fields).Error(errMsg)
}

// LogErrorConsumer for consumer
func LogErrorConsumer(errMsg string) {
	log := logrus.New()
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(os.Getenv("ELASTICSEARCH_HOST")))
	if err != nil {
		log.Println(err)
		log.Println(errMsg)
		return
	}

	hook, err := elogrus.NewAsyncElasticHook(client, os.Getenv("ELASTICSEARCH_HOST"), logrus.DebugLevel, "service-consumer-logger")
	if err != nil {
		log.Println(err)
		log.Println(errMsg)
	}
	log.Hooks.Add(hook)

	start := time.Now()

	latency := time.Since(start)

	log.WithFields(logrus.Fields{
		"Key":          uuid.GetUUID(),
		"ServiceName":  os.Getenv("SERVICE_NAME"),
		"ResponseTime": latency,
	}).Error(errMsg)
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
