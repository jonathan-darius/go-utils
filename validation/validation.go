package validation

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/forkyid/go-utils/rest"
	"github.com/forkyid/go-utils/rest/restid"
	"github.com/go-playground/validator"
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

func validateID(details *rest.ErrorDetails, id restid.ID, name string, tags []string) (code int) {
	code = http.StatusOK
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

		case "valid":
			if id.Encrypted != "" && !id.Valid {
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

	return code
}

func validateStructID(data interface{}) (details *rest.ErrorDetails, code int) {
	details = &rest.ErrorDetails{}
	code = http.StatusOK

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)
	id := restid.ID{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldT := t.Field(i)

		if field.Type().ConvertibleTo(reflect.TypeOf(id)) {
			name := strings.SplitN(fieldT.Tag.Get("json"), ",", 2)[0]
			if name == "" || name == "-" {
				name = strings.ToLower(fieldT.Name)
			}

			tags := strings.Split(fieldT.Tag.Get("id"), ",")
			id = field.Interface().(restid.ID)

			vCode := validateID(details, id, name, tags)
			if vCode != http.StatusOK && code != http.StatusBadRequest {
				code = vCode
			}
			continue
		}

		if field.Type().ConvertibleTo(reflect.TypeOf([]restid.ID{})) {
			name := strings.SplitN(fieldT.Tag.Get("json"), ",", 2)[0]
			if name == "" || name == "-" {
				name = strings.ToLower(fieldT.Name)
			}

			tags := strings.Split(fieldT.Tag.Get("id"), ",")
			ids := field.Interface().([]restid.ID)

			for i := range ids {
				vCode := validateID(details, ids[i], fmt.Sprintf("%v[%v]", name, i), tags)
				if vCode != http.StatusOK && code != http.StatusBadRequest {
					code = vCode
				}
			}
			continue
		}
	}

	return details, code
}

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
		for _, err := range err.(validator.ValidationErrors) {
			details.Add(err.Field(), err.Tag())
		}
		code = http.StatusBadRequest
	}

	idDet, idCode := validateStructID(data)
	if idCode != http.StatusOK {
		if code != http.StatusBadRequest {
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

	if code == http.StatusOK {
		details = nil
	}

	return details, code
}
