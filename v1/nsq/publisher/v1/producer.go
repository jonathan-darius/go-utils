package v1

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
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

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}

func drainBody(resp *http.Response) {
	if resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func pubTypeHTTPS(topic string, data []byte) (err error) {
	host := fmt.Sprintf("%s/pub?topic=%s", os.Getenv("NSQD_HOST"), topic)
	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodPost,
		host,
		bytes.NewReader(data),
	)
	var resp *http.Response
	for pubRetry := 0; pubRetry <= retry; pubRetry++ {
		resp, err = client.Do(req)
		if err != nil {
			if pubRetry >= retry {
				log.Printf("[ERROR] [NSQD] [%s] [%s] %v \n", host, err.Error(), string(data))
				break
			}
			time.Sleep(backoff(pubRetry))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			drainBody(resp)
			if pubRetry >= retry {
				log.Printf("[ERROR] [NSQD] [Code: %d] [%s] %v\n", resp.StatusCode, host, string(data))
				break
			}
			time.Sleep(backoff(pubRetry))
			continue
		}

		drainBody(resp)
		break
	}
	return
}
