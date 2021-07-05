package v1

import (
	"os"

	"github.com/joho/godotenv"
)

var _ = godotenv.Load()

var LogRoute = Route{
	ExchangeName: os.Getenv("RMQ_AI_LOG_EXCHANGE"),
	ExchangeType: "topic",
	RoutingKey:   os.Getenv("RMQ_AI_LOG_ROUTING_KEY"),
	QueueName:    os.Getenv("RMQ_AI_LOG_QUEUE"),
}