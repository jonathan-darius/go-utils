package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// Get params
// @key: string
// return interface
func Get(key string, seconds ...int) (interface{}, error) {
	if isCacheConnected() == false {
		return nil, fmt.Errorf("get: connect failed: %s", os.Getenv("REDIS_HOST"))
	}

	client := getRedisClient()

	var data interface{}

	resp := client.Get(key)
	if resp.Err() == redis.Nil {
		return nil, nil
	}

	if resp.Err() != nil {
		return nil, fmt.Errorf("get: redis get failed: %s: %s", os.Getenv("REDIS_HOST"), resp.Err().Error())
	}

	err := json.Unmarshal([]byte(resp.Val()), &data)
	if err != nil {
		return nil, fmt.Errorf("get: unmarshal failed: %s: %s", os.Getenv("REDIS_HOST"), err.Error())
	}

	if len(seconds) == 1 {
		go SetExpire(key, seconds[0])
	}

	return data, nil
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

func getKey(data interface{}) (key string, err error) {
	v := reflect.ValueOf(data)

	// check for nil and pointer dereference
	if data == nil {
		return
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldT := t.Field(i)

		// check tag exist
		cacheTag := fieldT.Tag.Get("cache")
		if len(cacheTag) == 0 {
			continue
		}
		tags := strings.Split(cacheTag, ",")

		// check empty value
		if tags[0] != "optional" && reflect.DeepEqual(field.Interface(), reflect.Zero(fieldT.Type).Interface()) {
			return "", fmt.Errorf("redis key: data cannot be empty")
		}

		// get json tag, else name
		name := strings.SplitN(fieldT.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			name = strings.ToLower(fieldT.Name)
		}

		// pointer dereference
		if field.Kind() == reflect.Ptr {
			field = v.Elem()
		}

		value := field.Interface()

		// if nested struct
		if tags[0] != "nodive" {
			if field.Kind() == reflect.Struct {
				value, err = getKey(value)
				if err != nil {
					return "", err
				}
			}
		}

		key = fmt.Sprintf("%v#%v:%v", key, name, value)
	}
	return key, nil
}

// Key params
// @data: interface{}
// @prefixes: ...string
// return string, error
func Key(data interface{}, prefixes ...string) (key string) {
	v := reflect.ValueOf(data)
	serviceName := os.Getenv("SERVICE_NAME")

	if serviceName == "" {
		log.Println("redis key: SERVICE_NAME env variable should not be empty")
		return ""
	}

	// for non struct based key
	if data == nil {
		key = serviceName

		for _, p := range prefixes {
			key = fmt.Sprintf("%v#%v", key, p)
		}
		return key
	}

	if reflect.DeepEqual(data, reflect.Zero(reflect.TypeOf(data)).Interface()) {
		log.Println("redis key: data should not be empty")
		return ""
	}

	// pointer dereference
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	key = fmt.Sprintf("%v#%v", serviceName, v.Type().Name())

	for _, p := range prefixes {
		key = fmt.Sprintf("%v#%v", key, p)
	}

	dataKey, err := getKey(data)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	key += dataKey

	return key
}
