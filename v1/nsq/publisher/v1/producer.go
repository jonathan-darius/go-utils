package v1

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	connection "github.com/forkyid/go-utils/v1/nsq"
	"golang.org/x/sync/semaphore"
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
	go publishHTTPS(data, topic, pubPool)
	return
}

func publishHTTPS(data []byte, topic string, pool *semaphore.Weighted) {
	pubRetry := retry
	host := fmt.Sprintf("%s/pub?topic=%s", os.Getenv("NSQD_HOST"), topic)
	defer pool.Release(1)
	req, _ := http.NewRequest(
		http.MethodPost,
		host,
		bytes.NewReader(data),
	)
	client := &http.Client{}
retPub:
	pubRetry--
	resp, err := client.Do(req)
	if err != nil {
		if pubRetry < 0 {
			log.Printf("[ERROR] [%s] [%s] %v \n", host, err.Error(), string(data))
			return
		}
		time.Sleep(3 * time.Second)
		goto retPub
	}
	if resp.StatusCode != http.StatusOK {
		if pubRetry < 0 {
			log.Printf("[ERROR] [%d] [%s] %v", resp.StatusCode, host, string(data))
			return
		}
		time.Sleep(3 * time.Second)
		goto retPub
	}
}
