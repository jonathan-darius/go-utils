package v1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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
	return
}

func pubTypeNSQD(topic string, data []byte) (err error) {
	producer, err := connection.Start()
	if err != nil {
		log.Println("failed to connect to nsqd: ", err.Error())
		return
	}

	err = producer.PublishAsync(topic, data, nil)
	if err != nil {
		log.Println("failed to publish data to nsqd: ", err.Error())
		return
	}
	return
}

func pubTypeHTTPS(topic string, data []byte) (err error) {
	pubPool := connection.StartProducerPool()
	ctx := context.Background()
	pubPool.Acquire(ctx, 1)
	go func() {
		defer pubPool.Release(1)
		publishHTTPS(data, topic)
	}()
	return
}

func publishHTTPS(data []byte, topic string) {
	pubRetry := retry
	host := fmt.Sprintf("%s/pub?topic=%s", os.Getenv("NSQD_HOST"), topic)
	req, _ := http.NewRequest(
		http.MethodPost,
		host,
		bytes.NewReader(data),
	)
	client := &http.Transport{}
retPub:
	pubRetry--
	resp, err := client.RoundTrip(req)
	if err != nil {
		if pubRetry < 0 {
			log.Printf("[ERROR] [%s] [%s] %v \n", host, err.Error(), string(data))
			return
		}
		time.Sleep(3 * time.Second)
		goto retPub
	}
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if pubRetry < 0 {
			log.Printf("[ERROR] [%d] [%s] %v", resp.StatusCode, host, string(data))
			return
		}
		time.Sleep(3 * time.Second)
		goto retPub
	}
}
