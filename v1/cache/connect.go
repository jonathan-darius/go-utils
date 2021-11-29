package cache

import (
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

var (
	redisClient *redis.Client
)

// ConnectCache func
func ConnectCache() {
	connPort := buildRedisEnvString()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     connPort,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
}

// IsCacheConnected func
// return bool
func IsCacheConnected() (isAlive bool) {
	if redisClient == nil {
		ConnectCache()
	}

	return redisClient.Ping().Val() == "PONG"
}

// getRedisClient func
// return *redis.Client
func getRedisClient() *redis.Client {
	if redisClient == nil {
		ConnectCache()
	}
	return redisClient
}

// buildRedisEnvString func
// return string
func buildRedisEnvString() (rtnString string) {
	if os.Getenv("REDIS_PORT") == "" {
		os.Setenv("REDIS_PORT", "6379")
	}

	if os.Getenv("REDIS_HOST") == "" {
		rtnString = fmt.Sprintf(":%s", os.Getenv("REDIS_PORT"))
		return
	}

	rtnString = fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	return
}
