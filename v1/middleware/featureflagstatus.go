package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/forkyid/go-utils/v1/logger"
	"github.com/forkyid/go-utils/v1/rest"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	DisabledStatus    = "disabled"
	MaintenanceStatus = "maintenance"
	ErrDisabled       = errors.New("Feature Is Disabled")
	ErrMaintenance    = errors.New("Feature Under Maintenance")
)

func (mid *Middleware) CheckFeatureFlagStatus(ctx *gin.Context, key string) {
	status, err := getFeatureFlagStatus(key)
	if err != nil {
		logger.Errorf(ctx, "get feature flag status", err)
		rest.ResponseMessage(ctx, http.StatusInternalServerError)
		ctx.Abort()
		return
	}
	if status == DisabledStatus {
		rest.ResponseMessage(ctx, http.StatusForbidden, ErrDisabled.Error())
		ctx.Abort()
		return
	}

	if status == MaintenanceStatus {
		rest.ResponseMessage(ctx, http.StatusForbidden, ErrMaintenance.Error())
		ctx.Abort()
		return
	}

	ctx.Next()
}

func getFeatureFlagStatus(key string) (status string, err error) {
	req := rest.Request{
		URL:    fmt.Sprintf("%v/cms/v1/feature-flag?key=%v", os.Getenv("API_ORIGIN_URL"), key),
		Method: http.MethodGet,
	}

	body, code := req.Send()
	if code != http.StatusOK {
		err = fmt.Errorf("[%v] %v: %v", req.Method, req.URL, string(body))
		return
	}

	data, err := rest.GetData(body)
	if err != nil {
		err = errors.Wrap(err, "get data")
		return
	}

	err = json.Unmarshal(data, &status)
	err = errors.Wrap(err, "unmarshall data")
	return
}
