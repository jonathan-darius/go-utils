package v1

import (
	"log"
	"os"

	connection "github.com/forkyid/go-utils/v1/nsq"
)

func Publish(data []byte) error {
	producer, err := connection.Start()
	if err != nil {
		log.Println("failed to connect to nsqd: ", err.Error())
		return err
	}

	err = producer.PublishAsync(os.Getenv("NSQD_TOPIC"), data, nil)
	if err != nil {
		log.Println("failed to publish data to nsqd: ", err.Error())
		return err
	}

	return err
}