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

func validateID(details *rest.ErrorDetails, id *restid.ID, name string, tags []string) (code int) {
	code = http.StatusOK
	allowZero := false
	for i := range tags {
		switch tags[i] {
		case "required":
			if id == nil || (*id).Encrypted == "" {
				details.Add(name, "required")
				if code != http.StatusBadRequest {
					code = http.StatusBadRequest
				}
				break
			}

			if !(*id).Valid {
				details.Add(name, "invalid")
				if code != http.StatusBadRequest {
					code = http.StatusUnprocessableEntity
				}
			}
			break

		case "valid":
			if id == nil || (*id).Encrypted != "" && !(*id).Valid {
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

	if !allowZero && id != nil && (*id).Valid && (*id).Raw == 0 {
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

		name := strings.SplitN(fieldT.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			name = strings.ToLower(fieldT.Name)
		}

		tags := strings.Split(fieldT.Tag.Get("id"), ",")

		if len(tags) <= 0 {
			continue
		}

		vCode := http.StatusOK
		switch field.Interface().(type) {
		case *restid.ID, nil:
			id := field.Interface().(*restid.ID)
			vCode = validateID(details, id, name, tags)
		case restid.ID:
			id = field.Interface().(restid.ID)
			vCode = validateID(details, &id, name, tags)
		case string:
			id = restid.IDFromEncrypted(field.Interface().(string))
			vCode = validateID(details, &id, name, tags)
		}
		if tags[0] == "dive" {
			switch field.Interface().(type) {
			case *[]restid.ID:
				ids := field.Interface().(*[]restid.ID)
				for i := range *ids {
					vCodei := validateID(details, &(*ids)[i], fmt.Sprintf("%v[%v]", name, i), tags)
					if vCode != http.StatusOK && vCode != http.StatusBadRequest {
						vCode = vCodei
					}
				}
			case []restid.ID:
				ids := field.Interface().([]restid.ID)
				for i := range ids {
					vCodei := validateID(details, &ids[i], fmt.Sprintf("%v[%v]", name, i), tags)
					if vCode != http.StatusOK && code != http.StatusBadRequest {
						vCode = vCodei
					}
				}
			case map[restid.ID]interface{}:
				ids := field.Interface().(map[restid.ID]interface{})
				for id := range ids {
					vCodei := validateID(details, &id, fmt.Sprintf("%v[%v]", name, id), tags)
					if vCode != http.StatusOK && code != http.StatusBadRequest {
						vCode = vCodei
					}
				}
			case interface{}:
				if field.Kind() == reflect.Slice {
					for i := 0; i < field.Len(); i++ {
						vDet, vCodei := validateStructID(field.Index(i).Interface())
						vCode = vCodei
						for k, v := range *vDet {
							details.Add(fmt.Sprintf("%v[%v].%v", name, i, k), v)
						}
					}
					break
				}
				vDet, vCodei := validateStructID(field.Interface())
				vCode = vCodei
				for k, v := range *vDet {
					details.Add(name+"."+k, v)
				}
			}
		}
		if vCode != http.StatusOK && code != http.StatusBadRequest {
			code = vCode
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
		if errV, ok := err.(validator.ValidationErrors); ok {
			for _, err := range errV {
				details.Add(err.Field(), err.Tag())
			}
			code = http.StatusBadRequest
		}
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
