package v1

import (
	"os"

	"github.com/joho/godotenv"
)

var _ = godotenv.Load()

var LogRoute = Route{
	ExchangeName: os.Getenv("RABBITMQ_LOG_EXCHANGE"),
	ExchangeType: "topic",
	RoutingKey:   os.Getenv("RABBITMQ_LOG_ROUTING_KEY"),
	QueueName:    os.Getenv("RABBITMQ_LOG_QUEUE"),
}