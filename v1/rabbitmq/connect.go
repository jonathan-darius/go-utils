package rabbitmq

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

func Start(m *sync.Mutex) (*amqp.Channel, error) {
	m.Lock()
	defer m.Unlock()

	var err error
	if conn == nil {
		if os.Getenv("RABBITMQ_PORT") == "" {
			os.Setenv("RABBITMQ_PORT", "5672")
		}

		connString := fmt.Sprintf("amqp://%s:%s@%s:%s",
			os.Getenv("RABBITMQ_USER"),
			os.Getenv("RABBITMQ_PASSWORD"),
			os.Getenv("RABBITMQ_HOST"),
			os.Getenv("RABBITMQ_PORT"),
		)

		conn, err = amqp.Dial(connString)
		if err != nil {
			log.Println(fmt.Sprintf("%s: %s", "Failed to connect to RabbitMQ", err.Error()))
			return nil, err
		}

		channel, err = conn.Channel()
		if err != nil {
			log.Println("Failed to open RabbitMQ channel:", err.Error())
			return nil, err
		}

		go func() {
			errChan := conn.NotifyClose(make(chan *amqp.Error))
			for {
				if <-errChan != nil {
					log.Println("RabbitMQ connection closed")
					conn = nil
					return
				}
			}
		}()

		go func() {
			errChan := channel.NotifyClose(make(chan *amqp.Error))
			for {
				if <-errChan != nil {
					log.Println("RabbitMQ channel closed")
					conn = nil
					return
				}
			}
		}()

		fmt.Println("Successfully connected to RabbitMQ")

	}

	return channel, err
}