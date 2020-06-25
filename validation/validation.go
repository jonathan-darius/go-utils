package validation

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/forkyid/go-utils/rest/restid"
	"github.com/go-playground/validator"
)

// ErrorDetails contains '|' separated details for each field
type ErrorDetails map[string]string

// Validator validator
var Validator = validator.New()

// Add adds details to key separated by '|'
func (details *ErrorDetails) Add(key, val string) {
	if (*details)[key] != "" {
		(*details)[key] += " | "
	}
	(*details)[key] += val
}

func validateID(data interface{}) (details *ErrorDetails, code int) {
	details = &ErrorDetails{}
	code = http.StatusOK

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)
	id := restid.ID{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldT := t.Field(i)
		if !field.Type().ConvertibleTo(reflect.TypeOf(id)) {
			continue
		}

		name := strings.SplitN(fieldT.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			name = strings.ToLower(fieldT.Name)
		}

		tags := strings.Split(fieldT.Tag.Get("id"), ",")
		id = field.Interface().(restid.ID)

		allowZero := false
		for i := range tags {
			switch tags[i] {
			case "required":
				if id.Encrypted == "" {
					details.Add(name, "required")
					if code != http.StatusBadRequest {
						code = http.StatusBadRequest
					}
					break
				}

				if !id.Valid {
					details.Add(name, "invalid")
					if code != http.StatusBadRequest {
						code = http.StatusUnprocessableEntity
					}
				}
				break

			case "allow-zero":
				allowZero = true
				break
			}
		}

		if !allowZero && id.Valid && id.Raw == 0 {
			details.Add(name, "invalid")
			if code != http.StatusBadRequest {
				code = http.StatusUnprocessableEntity
			}
		}
	}

	return details, code
}

func validateProcessable(data interface{}) (details *ErrorDetails, code int) {
	details = &ErrorDetails{}
	code = http.StatusOK

	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		fieldT := t.Field(i)

		name := strings.SplitN(fieldT.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			name = strings.ToLower(fieldT.Name)
		}

		tags := fieldT.Tag.Get("process")
		err := Validator.Var(field.Interface(), tags)
		if err != nil {
			for _, err := range err.(validator.ValidationErrors) {
				details.Add(name, err.Tag())
			}
			code = http.StatusUnprocessableEntity
		}
	}

	return details, code
}

// Validate handles common request errors
// returns error details and status code
func Validate(data interface{}) (details *ErrorDetails, code int) {
	details = &ErrorDetails{}
	code = http.StatusOK

	err := Validator.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			details.Add(err.Field(), err.Tag())
		}
		code = http.StatusBadRequest
	}

	if code != http.StatusBadRequest {
		idDet, idCode := validateID(data)
		if idCode != http.StatusOK {
			code = idCode
			for field, det := range *idDet {
				details.Add(field, det)
			}
		}
	}

	prDet, prCode := validateProcessable(data)
	if code != http.StatusBadRequest {
		if prCode != http.StatusOK {
			code = prCode
			for field, det := range *prDet {
				details.Add(field, det)
			}
		}
	}

	return details, code
}
