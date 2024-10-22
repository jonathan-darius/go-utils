package v1

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/useinsider/go-pkg/insrequester"

	connection "github.com/forkyid/go-utils/v1/nsq"
)

const (
	PubTypeNSQD  = "nsqd"
	PubTypeHTTPS = "https"
	retry        = 3
)

func Publish(data []byte) (err error) {
	publishType := os.Getenv("NSQD_PUB_TYPE")
	topic := os.Getenv("NSQD_TOPIC")

	switch publishType {
	case PubTypeHTTPS:
		err = pubTypeHTTPS(topic, data)
	default:
		err = pubTypeNSQD(topic, data)
	}
	return err
}

func pubTypeNSQD(topic string, data []byte) (err error) {
	producer, err := connection.Start()
	if err != nil {
		log.Println("failed to connect to nsqd: ", err.Error())
		return err
	}

	err = producer.PublishAsync(topic, data, nil)
	if err != nil {
		log.Println("failed to publish data to nsqd: ", err.Error())
		return
	}
	return
}

func pubTypeHTTPS(topic string, data []byte) (err error) {
	host := fmt.Sprintf("%s/pub?topic=%s", os.Getenv("NSQD_HOST"), topic)
	requester := insrequester.NewRequester().
		WithRetry(insrequester.RetryConfig{
			WaitBase: 2 * time.Second,
			Times:    retry,
		}).Load()
	resp, err := requester.Post(insrequester.RequestEntity{
		Endpoint: host,
		Body:     data,
	})
	if err != nil {
		log.Printf("[ERROR] [NSQD] [%s] [%s] %v \n", host, err.Error(), string(data))
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return
}
