package cache

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

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

func key(serviceName string, data interface{}, prefixes ...string) (key string, err error) {
	v := reflect.ValueOf(data)

	// for non struct based key
	if data == nil {
		key = serviceName

		for _, p := range prefixes {
			key = fmt.Sprintf("%v#%v", key, p)
		}
		return key, nil
	}

	if reflect.DeepEqual(data, reflect.Zero(reflect.TypeOf(data)).Interface()) {
		return "", fmt.Errorf("redis key: data should not be empty")
	}

	// pointer dereference
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("redis key: data should be a struct")
	}

	key = fmt.Sprintf("%v#%v", serviceName, v.Type().Name())

	for _, p := range prefixes {
		key = fmt.Sprintf("%v#%v", key, p)
	}

	dataKey, err := getKey(data)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	key += dataKey

	return key, nil
}

// Key params
// @data: interface{}
// @prefixes: ...string
// return string, error
func Key(data interface{}, prefixes ...string) string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		log.Println("redis key: SERVICE_NAME env variable should not be empty")
		return ""
	}

	key, err := key(serviceName, data, prefixes...)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return key
}

// ExternalKey params
// @serviceName: string
// @data: interface{}
// @prefixes: ...string
// return string, error
func ExternalKey(serviceName string, data interface{}, prefixes ...string) string {
	key, err := key(serviceName, data, prefixes...)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return key
}
