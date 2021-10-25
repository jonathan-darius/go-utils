package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"reflect"

	"github.com/gin-gonic/gin"
)

type File *multipart.FileHeader
type Files []*multipart.FileHeader

// BindJSON params
// 	@ctx: *gin.Context
// 	@v: interface{}
// 	return error
func BindJSON(ctx *gin.Context, v interface{}) (err error) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return
	}
	ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	err = json.Unmarshal(body, &v)
	return
}

// BindQuery params
// 	@ctx: *gin.Context
// 	@v: interface{}
// 	return error
func BindQuery(ctx *gin.Context, v interface{}) (err error) {
	return ctx.BindQuery(v)
}

// @BindFormData params
// 	@ctx: *gin.Context
// 	@v: interface{}
// 	return error
func BindFormData(ctx *gin.Context, v interface{}) (err error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		val = val.Elem()
	} else {
		err = errors.New("not a struct")
		return
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		tag := t.Field(i).Tag.Get("form")
		fieldType := val.Field(i).Type()

		if fieldType == reflect.TypeOf("") {
			val.Field(i).SetString(
				ctx.PostForm(tag),
			)
		} else if fieldType == reflect.TypeOf([]string{}) {
			val.Field(i).Set(
				reflect.ValueOf(ctx.PostFormArray(tag)),
			)
		}

	}

	v = val
	return ctx.ShouldBind(&v)
}

// BindMultipartFormData params
// 	@ctx: *gin.Context
// 	@v: interface{}
// 	return error
func BindMultipartFormData(ctx *gin.Context, v interface{}) (err error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		val = val.Elem()
	} else {
		err = errors.New("not a struct")
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		return
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		tagForm := t.Field(i).Tag.Get("form")
		tagFormFile := t.Field(i).Tag.Get("form-file")
		fieldType := val.Field(i).Type()

		if fieldType == reflect.TypeOf([]string{}) {
			val.Field(i).Set(reflect.ValueOf(form.Value[tagForm]))
		} else if fieldType == reflect.TypeOf("") && len(form.Value[tagForm]) > 0 {
			val.Field(i).SetString(form.Value[tagForm][0])
		}

		var file File
		if fieldType == reflect.TypeOf(Files{}) {
			val.Field(i).Set(reflect.ValueOf(form.File[tagFormFile]))
		} else if fieldType == reflect.TypeOf(file) && len(form.File[tagFormFile]) > 0 {
			val.Field(i).Set(reflect.ValueOf(form.File[tagFormFile][0]))
		}
	}

	v = val
	return
}
