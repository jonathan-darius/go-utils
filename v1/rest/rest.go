package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/forkyid/go-utils/v1/logger"
	publisher "github.com/forkyid/go-utils/v1/rabbitmq/publisher/v1"
	responseMsg "github.com/forkyid/go-utils/v1/rest/response"
	uuid "github.com/forkyid/go-utils/v1/uuid"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/go-playground/validator/v10"
)

// Response types
type Response struct {
	Body    interface{}       `json:"body,omitempty"`
	Error   string            `json:"error,omitempty"`
	Message string            `json:"message,omitempty"`
	Detail  map[string]string `json:"detail,omitempty"`
	Status  int               `json:"status,omitempty"`
}

// ResponseResult types
type ResponseResult struct {
	Context *gin.Context
	UUID    string
}

// Log uses current response context to log
func (resp ResponseResult) Log(message string) {
	logger.LogWithContext(resp.Context, resp.UUID, message)
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

	var copied gin.Context = *context
	go PublishLog(&copied, status, payload, msg[0])
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
	
	var copied gin.Context = *context
	go PublishLog(&copied, status, nil, msg[0])
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

// PublishLog params
// 	@context: *gin.Context
// 	@status: int
// 	@payload: interface
// 	@msg: []string
//	@return error
func PublishLog(context *gin.Context, status int, payload interface{}, msg ...string) error {
	requestBody, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		log.Println("read body failed " + err.Error())
		return nil
	}

	var requestBodyInterface map[string]interface{}
	if len(requestBody) > 1 {
		err = json.Unmarshal(requestBody, &requestBodyInterface)
		if err != nil {
			log.Println("unmarshal data failed " + err.Error())
			return nil
		}
	}

	body := &LogData{
		Request: LogRequest{
			Method:          context.Request.Method,
			URL:             context.Request.URL,
			Header:          context.Request.Header,
			Body:            requestBodyInterface,
			Host:            context.Request.Host,
			Form:            context.Request.Form,
			PostForm:        context.Request.PostForm,
			MultipartForm:   context.Request.MultipartForm,
			RemoteAddr:      context.Request.RemoteAddr,
			PublicIPAddress: context.ClientIP(),
			RequestURI:      context.Request.RequestURI,
		},
		Response: Response{
			Body:    payload,
			Status:  status,
			Message: msg[0],
		},
	}

	location, err := time.LoadLocation("UTC")
	if err != nil {
		log.Println("failed on get location: " + err.Error())
		return nil
	}

	utcTime := time.Now().In(location).Format(time.RFC3339Nano)
	data, err := json.Marshal(map[string]interface{}{
		"SERVICE_NAME": os.Getenv("SERVICE_NAME"),
		"payload": body,
		"timestamp": utcTime,
	})
	if err != nil {
		log.Println("failed on encoding json: " + err.Error())
		return nil
	}

	err = publisher.LogRoute.Publish(&publisher.Publish{
		Body: string(data),
	})
	if err != nil {
		log.Println("publish data failed " + err.Error())
		return nil
	}

	return nil
}
