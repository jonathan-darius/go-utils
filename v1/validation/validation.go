package validation

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/forkyid/go-utils/v1/rest"
	"github.com/go-playground/validator/v10"
)

// Validator validator
var Validator = func() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}()

func validateProcessable(data interface{}) (details *rest.ErrorDetails, code int) {
	details = &rest.ErrorDetails{}
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
func Validate(data interface{}) (details *rest.ErrorDetails, code int) {
	details = &rest.ErrorDetails{}
	code = http.StatusOK

	err := Validator.Struct(data)
	if err != nil {
		if errV, ok := err.(validator.ValidationErrors); ok {
			for _, err := range errV {
				details.Add(err.Field(), err.Tag())
			}
			code = http.StatusBadRequest
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

	if code == http.StatusOK {
		details = nil
	}

	return details, code
}
