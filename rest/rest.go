package rest

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/forkyid/go-utils/logger"
	uuid "github.com/forkyid/go-utils/uuid"
	"github.com/forkyid/go-utils/rest/constants"
	"github.com/gin-gonic/gin"
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

func (resp ResponseResult) Log(message string) {
	logger.LogError(resp.Context, resp.UUID, message)
}

// ResponseData params
// @context: *gin.Context
// status: int
// msg: string
func ResponseData(context *gin.Context, status int, payload interface{}, msg ...string) ResponseResult {
	if len(msg) > 1 {
		log.Panicln("response cannot contain more than one message")
	}
	if len(msg) == 0 {
		if defaultMessage := constants.Response[status]; defaultMessage == nil {
			log.Panicln("default message for status code " + strconv.Itoa(status) + " not found")
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
		log.Panicln("response cannot contain more than one message")
	}
	if len(msg) == 0 {
		if defaultMessage := constants.Response[status]; defaultMessage == nil {
			log.Panicln("default message for status code " + strconv.Itoa(status) + " not found")
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
		log.Panicln("response cannot contain more than one message")
	}
	if len(msg) == 0 {
		if defaultMessage := constants.Response[status]; defaultMessage == nil {
			log.Panicln("default message for status code " + strconv.Itoa(status) + " not found")
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
	} else if det, ok := detail.(string); ok {
		response.Detail = map[string]string{}
		response.Detail["error"] = det
	}

	context.JSON(status, response)
	return ResponseResult{context, response.Error}
}

func PostPayload(url, payload string, headers map[string]string) ([]byte, int) {
	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR: [POST] " + url + " " + err.Error())
		return nil, 0
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, resp.StatusCode
}

func GetPayload(url string, headers map[string]string, reqBody io.Reader, wg *sync.WaitGroup, responseBody *[][]byte, responseCode *[]int) ([]byte, int) {
	req, _ := http.NewRequest("GET", url, reqBody)

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR: [GET]", err.Error())
		return nil, -1
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if responseBody != nil {
		*responseBody = append(*responseBody, body)
	}

	if responseCode != nil {
		*responseCode = append(*responseCode, resp.StatusCode)
	}

	if wg != nil {
		wg.Done()
	}

	return body, resp.StatusCode
}
