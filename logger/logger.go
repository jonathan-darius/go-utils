package logger

import (
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

func Init() {
	var Environment = os.Getenv("ENV")
	if Environment == "production" {
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

	log.WithFields(logrus.Fields{
		"Key":           uuid,
		"ServiceName":   os.Getenv("SERVICE_NAME"),
		"Request":       r.RequestURI,
		"Method":        r.Method,
		"IP":            realIP(r),
		"RemoteAddress": r.Header.Get("X-Request-Id"),
		"Payload":       payload,
		"StatusCode":    lw.statusCode,
		"ResponseTime":  latency,
	}).Error(errMsg)
}

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
		log.Panic(err)
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
