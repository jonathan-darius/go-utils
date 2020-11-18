package middleware

import (
	"fmt"
	"net/http"
	"os"

	"github.com/forkyid/go-utils/aes"
	"github.com/forkyid/go-utils/cache"
	"github.com/forkyid/go-utils/jwt"
	"github.com/forkyid/go-utils/rest"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

func Authorization(ctx *gin.Context) {
	id, err := jwt.ExtractID(ctx.GetHeader("Authorization"))
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	banned, err := isBanned(id, ctx.GetHeader("Authorization"))
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusInternalServerError).
			Log("auth: check banned: " + err.Error())
		ctx.Abort()
		return
	}
	if banned {
		rest.ResponseMessage(ctx, http.StatusForbidden, "Banned")
		ctx.Abort()
		return
	}

	suspended, err := isSuspended(aes.Encrypt(id))
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusInternalServerError).
			Log("auth: check suspend: " + err.Error())
		ctx.Abort()
		return
	}
	if suspended {
		rest.ResponseMessage(ctx, http.StatusForbidden, "Suspend")
		ctx.Abort()
		return
	}

	loggedIn, err := isLoggedIn(id, ctx.GetHeader("X-Unique-ID"))
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusInternalServerError).
			Log("auth: whitelist: redis check exists: " + err.Error())
		ctx.Abort()
		return
	}

	if !loggedIn {
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	ctx.Next()
}

type Ban struct {
	ID string `cache:"key"`
}

func isBanned(memberID int, bearer string) (bool, error) {
	redisKey := cache.ExternalKey("global", Ban{
		ID: aes.Encrypt(memberID),
	})

	banned, err := cache.IsCacheExists(redisKey)
	if err != nil {
		return false, errors.Wrap(err, "redis: check exists")
	}
	if banned {
		return true, nil
	}

	header := map[string]string{
		"Authorization": bearer,
	}

	query := map[string]string{
		"block_type_id": aes.Encrypt(1),
		"blocker_id":    "0",
		"blocked_id":    aes.Encrypt(memberID),
	}

	req := rest.Request{
		URL:     fmt.Sprintf("%v/report/v1/blocks", os.Getenv("API_ORIGIN_URL")),
		Method:  http.MethodGet,
		Headers: header,
		Queries: query,
	}
	_, code := req.Send()
	if code != http.StatusOK {
		if code != http.StatusNotFound {
			return false, fmt.Errorf("get blocked: status code unexpected: %d", code)
		}
		return false, nil
	}

	err = cache.SetJSON(redisKey, "", 600)
	if err != nil {
		return false, errors.Wrap(err, "redis: set cache")
	}

	return true, nil
}

type Suspend struct {
	ID string `cache:"key"`
}

func isSuspended(memberID string) (bool, error) {
	suspended, err := cache.IsCacheExists(
		cache.ExternalKey("global", Suspend{
			ID: memberID,
		}))
	return suspended, errors.Wrap(err, "redis: check exists")
}

type WhiteList struct {
	ID       string `cache:"key"`
	DeviceID string `cache:"optional" json:"device_id"`
}

func isLoggedIn(memberID int, deviceID string) (bool, error) {
	loggedIn, err := cache.IsCacheExists(
		cache.ExternalKey("global", WhiteList{
			ID:       aes.Encrypt(memberID),
			DeviceID: deviceID,
		}))
	if err != nil {
		return false, errors.Wrap(err, "redis: check exists")
	}
	return loggedIn, nil
}
