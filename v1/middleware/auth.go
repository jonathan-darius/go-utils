package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/forkyid/go-utils/v1/aes"
	"github.com/forkyid/go-utils/v1/cache"
	"github.com/forkyid/go-utils/v1/jwt"
	"github.com/forkyid/go-utils/v1/logger"
	"github.com/forkyid/go-utils/v1/rest"
	"github.com/forkyid/go-utils/v1/uuid"
	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

type MemberStatusKey struct {
	ID string `cache:"key"`
}

type MemberStatus struct {
	DeviceID   string     `json:"device_id,omitempty" mapstructure:"device_id"`
	IsBanned   bool       `json:"is_banned,omitempty" mapstructure:"is_banned"`
	SuspendEnd *time.Time `json:"suspend_end,omitempty"`
}

func GetStatus(ctx *gin.Context, es *elastic.Client, memberID int) (status MemberStatus, err error) {
	statusKey := cache.ExternalKey("global", MemberStatusKey{
		ID: aes.Encrypt(memberID),
	})

	err = cache.GetUnmarshal(statusKey, &status, 600)
	if err != nil && err != redis.Nil {
		logger.LogWithContext(ctx, uuid.GetUUID(), errors.Wrap(err, "redis").Error())
	}

	if err != nil || err == redis.Nil {
		status.IsBanned, err = isBanned(ctx)
		if err != nil {
			return status, errors.Wrap(err, "check banned")
		}

		err = cache.SetJSON(statusKey, status, 600)
		if err != nil {
			logger.LogWithContext(ctx, uuid.GetUUID(), errors.Wrap(err, "redis set").Error())
		}

		return status, nil
	}

	return status, nil
}

func (mid *Middleware) Auth(ctx *gin.Context) {
	id, err := jwt.ExtractID(ctx.GetHeader("Authorization"))
	if err != nil {
		rest.ResponseError(ctx, http.StatusUnauthorized, map[string]string{
			"access_token": "expired",
		})
		log.Println("access_token:", err.Error())
		ctx.Abort()
		return
	}

	status, err := GetStatus(ctx, mid.elastic, id)
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusInternalServerError).
			Log("auth: " + err.Error())
		ctx.Abort()
		return
	}

	if status.IsBanned {
		rest.ResponseMessage(ctx, http.StatusForbidden, "Banned")
		ctx.Abort()
		return
	}

	if status.SuspendEnd != nil && status.SuspendEnd.After(time.Now()) {
		rest.ResponseMessage(ctx, http.StatusLocked, "Suspended")
		ctx.Abort()
		return
	}

	deviceID := ctx.GetHeader("X-Unique-ID")
	if status.DeviceID != deviceID {
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	ctx.Next()
}

func isBanned(ctx *gin.Context) (bool, error) {
	id, _ := jwt.ExtractID(ctx.GetHeader("Authorization"))
	query := map[string]string{
		"block_type_id": aes.Encrypt(1),
		"blocker_id":    "0",
		"blocked_id":    aes.Encrypt(id),
	}
	req := rest.Request{
		URL:     fmt.Sprintf("%v/report/v1/blocks", os.Getenv("API_ORIGIN_URL")),
		Method:  http.MethodGet,
		Queries: query,
	}
	_, code := req.WithContext(ctx).Send()
	if code == http.StatusNotFound {
		return false, nil
	}
	if code != http.StatusOK {
		return false, fmt.Errorf("get blocked: status code unexpected: %d", code)
	}
	return true, nil
}

func (Middleware) IsSuspended(feature string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		client, _ := jwt.ExtractClient(ctx.GetHeader("Authorization"))
		username := client.Username

		isSuspended, err := cache.IsCacheExists(username + ":" + feature)
		if err != nil {
			log.Println("failed on getting suspend data from redis: " + err.Error())
		}
		if isSuspended {
			ttl, err := cache.TTL(username + ":" + feature)
			if err != nil {
				log.Println("failed on getting ttl from redis: " + err.Error())
			} else {
				rest.ResponseData(ctx, http.StatusLocked, map[string]interface{}{
					"until": time.Now().Add(time.Second * time.Duration(ttl)).Format(time.RFC3339),
				}, "Locked")
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
