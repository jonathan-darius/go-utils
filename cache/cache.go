package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/go-redis/redis"
)

// Get params
// @key: string
// return interface{}, error
func Get(key string, seconds ...int) (interface{}, error) {
	if isCacheConnected() == false {
		return nil, fmt.Errorf("get: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	resp := client.Get(key)
	if resp.Err() == redis.Nil {
		return nil, nil
	}

	if resp.Err() != nil {
		return nil, fmt.Errorf("get: redis get failed: %s: %s", os.Getenv("REDIS_HOST"), resp.Err().Error())
	}

	var data interface{}
	err := json.Unmarshal([]byte(resp.Val()), &data)
	if err != nil {
		return nil, fmt.Errorf("get: unmarshal failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}

	if len(seconds) == 1 {
		go SetExpire(key, seconds[0])
	}

	return data, nil
}

// GetUnmarshal params
// @key: string
// @target: interface{}
// return error
func GetUnmarshal(key string, target interface{}, seconds ...int) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		fmt.Println("unmarshal target is not a pointer")
	}
	if isCacheConnected() == false {
		return fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	resp := client.Get(key)
	if resp.Err() == redis.Nil {
		return nil
	}

	if resp.Err() != nil {
		return fmt.Errorf("redis get failed: %s: %s", os.Getenv("REDIS_HOST"), resp.Err().Error())
	}

	err := json.Unmarshal([]byte(resp.Val()), target)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}

	if len(seconds) == 1 {
		go SetExpire(key, seconds[0])
	}

	return nil
}

// SetJSON params
// @key: string
// @value: interface{}
// @seconds: int
// return error
func SetJSON(key string, value interface{}, seconds int) error {
	if isCacheConnected() == false {
		return fmt.Errorf("setjson: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("setjson: redis set failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}

	return client.Set(key, valueJSON, time.Duration(seconds)*time.Second).Err()
}

// IsCacheExists params
// @key: string
// return bool, error
func IsCacheExists(key string) (bool, error) {
	if isCacheConnected() == false {
		return false, fmt.Errorf("cache exists: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()
	res := client.Exists(key)
	if res.Err() != nil {
		return false, fmt.Errorf("cache exists: check failed: %s", res.Err().Error())
	}

	return res.Val() != 0, nil
}

// SetExpire params
// @key: string
// @seconds: int
// return error
func SetExpire(key string, seconds int) error {
	if isCacheConnected() == false {
		return fmt.Errorf("set expire: connect failed: %s", os.Getenv("REDIS_HOST"))
	}
	client := getRedisClient()

	if err := client.Expire(key, time.Duration(seconds)*time.Second).Err(); err != nil {
		return fmt.Errorf("set expire: set expire failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}
	return nil
}

// Delete params
// @key: string
// return error
func Delete(key ...string) error {
	if isCacheConnected() == false {
		return fmt.Errorf("delete: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	if err := client.Del(key...).Err(); err != nil {
		return fmt.Errorf("delete: delete failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}
	return nil
}

// Purge params
// @key: string
// return error
func Purge(key string) error {
	if isCacheConnected() == false {
		return fmt.Errorf("purge: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	cursor := client.Scan(0, "*"+key+"*", 0).Iterator()
	err := cursor.Err()
	if err != nil {
		return fmt.Errorf("purge: cursor scan failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}

	for cursor.Next() {
		err := client.Del(cursor.Val()).Err()
		if err != nil {
			return fmt.Errorf("purge: delete failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
		}
	}

	return nil
}

// TTL params
// @key: string
// return float64, error
func TTL(key string) (float64, error) {
	if isCacheConnected() == false {
		return 0, fmt.Errorf("ttl: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()
	if client == nil {
		return 0, fmt.Errorf("ttl: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	duration := client.TTL(key)
	res, err := duration.Val().Seconds(), duration.Err()
	if err != nil {
		return 0, fmt.Errorf("ttl: set duration failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}
	return res, nil
}
