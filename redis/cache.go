package redis

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/forkyid/go-utils/logger"
	"github.com/forkyid/go-utils/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// Get params
// @key: string
// return interface
func Get(key string, seconds ...int) interface{} {
	if isCacheConnected() == false {
		ctx, _ := gin.CreateTestContext(nil)
		logger.LogError(ctx, uuid.GetUUID(), "failed on connecting to redis: "+os.Getenv("REDIS_HOST")+": no such host")
		return nil
	}

	client := getRedisClient()

	var data interface{}

	resp := client.Get(key)
	if resp.Err() != nil && resp.Err() != redis.Nil {
		ctx, _ := gin.CreateTestContext(nil)
		logger.LogError(ctx, uuid.GetUUID(), "failed on getting data from redis: "+resp.Err().Error())
		return nil
	}

	err := json.Unmarshal([]byte(resp.Val()), &data)
	if err != nil {
		return nil
	}

	if len(seconds) == 1 {
		go SetExpire(key, seconds[0])
	}

	return data
}

// SetJSON params
// @key: string
// @value: interface{}
// @seconds: int
// return error
func SetJSON(key string, value interface{}, seconds int) error {
	if isCacheConnected() == false {
		ctx, _ := gin.CreateTestContext(nil)
		logger.LogError(ctx, uuid.GetUUID(), "failed on connecting to redis: "+os.Getenv("REDIS_HOST")+": no such host")
		return nil
	}

	client := getRedisClient()

	valueJSON, err := json.Marshal(value)
	if err != nil {
		ctx, _ := gin.CreateTestContext(nil)
		logger.LogError(ctx, uuid.GetUUID(), "failed on connecting to redis: "+os.Getenv("REDIS_HOST")+": no such host")
		return err
	}

	return client.Set(key, valueJSON, time.Duration(seconds)*time.Second).Err()
}

// IsCacheExists params
// @key: string
// return bool, error
func IsCacheExists(key string) (bool, error) {
	if isCacheConnected() == false {
		return false, fmt.Errorf("failed on connecting to redis: %s: no such host", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()
	res := client.Exists(key)
	if res.Err() != nil {
		return false, res.Err()
	}

	return res.Val() != 0, nil
}

// SetExpire params
// @key: string
// @seconds: int
// return error
func SetExpire(key string, seconds int) error {
	if isCacheConnected() == false {
		return fmt.Errorf("failed on connecting to redis: %s: no such host", os.Getenv("REDIS_HOST"))
	}
	client := getRedisClient()

	return client.Expire(key, time.Duration(seconds)*time.Second).Err()
}

// Delete params
// @key: string
// return error
func Delete(key ...string) error {
	if isCacheConnected() == false {
		return fmt.Errorf("failed on connecting to redis: %s: no such host", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	return client.Del(key...).Err()
}

// Purge params
// @key: string
// return error
func Purge(key string) error {
	if isCacheConnected() == false {
		return fmt.Errorf("failed on connecting to redis: %s: no such host", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	cursor := client.Scan(0, "*"+key+"*", 0).Iterator()
	err := cursor.Err()
	if err != nil {
		return err
	}

	for cursor.Next() {
		err := client.Del(cursor.Val()).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// TTL params
// @key: string
// return float64, error
func TTL(key string) (float64, error) {
	client := getRedisClient()
	if client == nil {
		return 0, fmt.Errorf("failed on connecting to redis: %s: no such host", os.Getenv("REDIS_HOST"))
	}

	duration := client.TTL(key)

	return duration.Val().Seconds(), duration.Err()
}
