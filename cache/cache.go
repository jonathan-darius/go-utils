package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// Get params
// @key: string
// return interface{}, error
func Get(key string, seconds ...int) (interface{}, error) {
	if isCacheConnected() == false {
		return nil, fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	resp := client.Get(key)
	if resp.Err() == redis.Nil {
		return nil, nil
	}

	if resp.Err() != nil {
		return nil, errors.Wrap(resp.Err(), "redis get")
	}

	var data interface{}
	err := json.Unmarshal([]byte(resp.Val()), &data)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal")
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
		return resp.Err()
	}

	if resp.Err() != nil {
		return errors.Wrap(resp.Err(), "redis get")
	}

	err := json.Unmarshal([]byte(resp.Val()), target)
	if err != nil {
		return errors.Wrap(err, "unmarshal")
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
		return fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "marshal failed")
	}

	return errors.Wrap(client.Set(key, valueJSON, time.Duration(seconds)*time.Second).Err(), "redis set failed")
}

// IsCacheExists params
// @key: string
// return bool, error
func IsCacheExists(key string) (bool, error) {
	if isCacheConnected() == false {
		return false, fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()
	res := client.Exists(key)
	if res.Err() != nil {
		return false, errors.Wrap(res.Err(), "redis check failed")
	}

	return res.Val() != 0, nil
}

// SetExpire params
// @key: string
// @seconds: int
// return error
func SetExpire(key string, seconds int) error {
	if isCacheConnected() == false {
		return fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}
	client := getRedisClient()

	if err := client.Expire(key, time.Duration(seconds)*time.Second).Err(); err != nil {
		return errors.Wrap(err, "set expire failed")
	}
	return nil
}

// Delete params
// @key: string
// return error
func Delete(key ...string) error {
	if isCacheConnected() == false {
		return fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	if err := client.Del(key...).Err(); err != nil {
		return errors.Wrap(err, "delete failed")
	}
	return nil
}

// Purge params
// @key: string
// return error
func Purge(key string) error {
	if isCacheConnected() == false {
		return fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	cursor := client.Scan(0, "*"+key+"*", 0).Iterator()
	err := cursor.Err()
	if err != nil {
		return errors.Wrap(err, "cursor scan failed")
	}

	for cursor.Next() {
		err := client.Del(cursor.Val()).Err()
		if err != nil {
			return errors.Wrap(err, "delete failed")
		}
	}

	return nil
}

// TTL params
// @key: string
// return float64, error
func TTL(key string) (float64, error) {
	if isCacheConnected() == false {
		return 0, fmt.Errorf("redis connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	duration := client.TTL(key)
	res, err := duration.Val().Seconds(), duration.Err()
	if err != nil {
		return 0, errors.Wrap(err, "set TTL failed")
	}
	return res, nil
}
