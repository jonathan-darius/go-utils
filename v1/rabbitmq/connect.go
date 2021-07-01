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

func Channel(m *sync.Mutex) (*amqp.Channel, error) {
	if conn == nil {
		var err error
		conn, err = Start(m)
		if err != nil {
			return nil, err
		}
	}
	m.Lock()

	if channel == nil {
		var err error
		channel, err = conn.Channel()
		if err != nil {
			return nil, err
		}
	}

	m.Unlock()
	return channel, nil
}

func Start(m *sync.Mutex) (*amqp.Connection, error) {
	m.Lock()
	defer m.Unlock()

	if conn == nil || conn.IsClosed() {
		var err error
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
		fmt.Println("successfully connected to RabbitMQ")
	}

	return conn, nil
}
