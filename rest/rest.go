package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/forkyid/go-utils/logger"
	responseMsg "github.com/forkyid/go-utils/rest/response"
	uuid "github.com/forkyid/go-utils/uuid"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/go-playground/validator"
)

// Response types
type Response struct {
	Body    interface{}       `json:"body,omitempty"`
	Error   string            `json:"error,omitempty"`
	Message string            `json:"message,omitempty"`
	Detail  map[string]string `json:"detail,omitempty"`
}

// ResponseResult types
type ResponseResult struct {
	Context *gin.Context
	UUID    string
}

// Log uses current response context to log
func (resp ResponseResult) Log(message string) {
	logger.LogError(resp.Context, resp.UUID, message)
}

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

// ResponseData params
// @context: *gin.Context
// status: int
// msg: string
func ResponseData(context *gin.Context, status int, payload interface{}, msg ...string) ResponseResult {
	if len(msg) > 1 {
		log.Println("response cannot contain more than one message")
		log.Println("proceeding with first message only...")
	}
	if len(msg) == 0 {
		if defaultMessage := responseMsg.Response[status]; defaultMessage == nil {
			log.Println("default message for status code " + strconv.Itoa(status) + " not found")
			log.Println("proceeding with empty message...")
			msg = []string{""}
		} else {
			msg = []string{defaultMessage.(string)}
		}
	}

	response := Response{
		Body:    payload,
		Message: msg[0],
	}

	context.JSON(status, response)
	return ResponseResult{context, uuid.GetUUID()}
}

// ResponseMessage params
// @context: *gin.Context
// status: int
// msg: string
func ResponseMessage(context *gin.Context, status int, msg ...string) ResponseResult {
	if len(msg) > 1 {
		log.Println("response cannot contain more than one message")
		log.Println("proceeding with first message only...")
	}
	if len(msg) == 0 {
		if defaultMessage := responseMsg.Response[status]; defaultMessage == nil {
			log.Println("default message for status code " + strconv.Itoa(status) + " not found")
			log.Println("proceeding with empty message...")
			msg = []string{""}
		} else {
			msg = []string{defaultMessage.(string)}
		}
	}

	response := Response{
		Message: msg[0],
	}
	if status < 200 || status > 299 {
		response.Error = uuid.GetUUID()
	}

	context.JSON(status, response)
	return ResponseResult{context, response.Error}
}

// ResponseError params
// @context: *gin.Context
// status: int
// msg: string
// detail: array
func ResponseError(context *gin.Context, status int, detail interface{}, msg ...string) ResponseResult {
	if len(msg) > 1 {
		log.Println("response cannot contain more than one message")
		log.Println("proceeding with first message only...")
	}
	if len(msg) == 0 {
		if defaultMessage := responseMsg.Response[status]; defaultMessage == nil {
			log.Println("default message for status code " + strconv.Itoa(status) + " not found")
			log.Println("proceeding with empty message...")
			msg = []string{""}
		} else {
			msg = []string{defaultMessage.(string)}
		}
	}

	response := Response{
		Error:   uuid.GetUUID(),
		Message: msg[0],
	}

	if det, ok := detail.(validator.ValidationErrors); ok {
		response.Detail = map[string]string{}
		for _, err := range det {
			response.Detail[strings.ToLower(err.Field())] = err.Tag()
		}
	} else if det, ok := detail.(map[string]string); ok {
		response.Detail = det
	} else if det, ok := detail.(*ErrorDetails); ok {
		response.Detail = *det
	} else if det, ok := detail.(string); ok {
		response.Detail = map[string]string{}
		response.Detail["error"] = det
	}

	context.JSON(status, response)
	return ResponseResult{context, response.Error}
}

// MultipartForm creates multipart payload
func MultipartForm(fileKey string, files [][]byte, params map[string]string, multiParams map[string][]string) (io.Reader, string) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for _, j := range files {
		part, _ := writer.CreateFormFile(fileKey, bson.NewObjectId().Hex())
		part.Write(j)
	}
	for k, v := range multiParams {
		for _, j := range v {
			writer.WriteField(k, j)
		}
	}
	for k, v := range params {
		writer.WriteField(k, v)
	}
	err := writer.Close()
	if err != nil {
		return nil, ""
	}

	return body, writer.FormDataContentType()
}

// GetData unwraps "body" object
func GetData(jsonBody []byte) (json.RawMessage, error) {
	body := map[string]json.RawMessage{}
	err := json.Unmarshal(jsonBody, &body)
	if err != nil {
		return nil, err
	}
	data := body["body"]
	return data, err
}
