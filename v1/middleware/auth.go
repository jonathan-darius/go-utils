package middleware

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/mitchellh/mapstructure"
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
		status, err := get(es, memberID)
		if err != nil {
			return status, errors.Wrap(err, "get member data from: es")
		}

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
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
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

func isSuspended(feature string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		client, _ := jwt.ExtractClient(ctx.GetHeader("Authorization"))
		username := client.MemberUsername

		isSuspended, err := cache.IsCacheExists(username + ":" + feature)
		if err != nil {
			logger.Log("failed on getting suspend data from redis: " + err.Error())
		}
		if isSuspended {
			ttl, err := cache.TTL(username + ":" + feature)
			if err != nil {
				logger.Log("failed on getting ttl from redis: " + err.Error())
			} else {
				rest.ResponseData(ctx, http.StatusLocked, map[string]interface{}{
					"until": time.Now().Add(time.Second * time.Duration(ttl)).Format("2006-01-02T15:04:05.999Z"),
				}, "Locked")
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}

func get(es *elastic.Client, id int) (status MemberStatus, err error) {
	query := elastic.NewMatchQuery("id", aes.Encrypt(id))
	searchResult, err := es.Search().
		Index("users").
		Type("_doc").
		Query(query).
		Do(context.Background())

	if err != nil {
		return status, errors.Wrap(err, "elastic")
	}

	if searchResult == nil || searchResult.TotalHits() == 0 {
		return status, errors.Wrap(err, "member not found")
	}

	user := map[string]interface{}{}
	err = json.Unmarshal(searchResult.Hits.Hits[0].Source, &user)
	if err != nil {
		return status, errors.Wrap(err, "unmarshal")
	}

	err = mapstructure.Decode(user["status"], &status)
	if err != nil {
		return status, errors.Wrap(err, "mapstructure.Decode")
	}

	return status, nil
}
