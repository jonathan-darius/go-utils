package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/forkyid/go-utils/aes"
	"github.com/forkyid/go-utils/cache"
	"github.com/forkyid/go-utils/jwt"
	"github.com/forkyid/go-utils/rest"
	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

type MemberStatus struct {
	ID string `cache:"key"`
}

type MemberData struct {
	DeviceID   string                 `json:"device_id,omitempty"`
	IsBanned   bool                   `json:"is_banned,omitempty"`
	SuspendEnd *time.Time             `json:"suspend_end,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

func GetMemberStatus(ctx *gin.Context, es *elastic.Client, memberID int) (status MemberData, err error) {
	statusKey := cache.ExternalKey("global", MemberStatus{
		ID: aes.Encrypt(memberID),
	})

	err = cache.GetUnmarshal(statusKey, &status, 600)
	if err != nil && err != redis.Nil {
		return status, errors.Wrap(err, "redis get")
	}

	if err == redis.Nil {
		status, err := getMemberData(es, memberID)
		if err != nil {
			return status, errors.Wrap(err, "get member data from: es")
		}

		status.IsBanned, err = isBanned(ctx)
		if err != nil {
			return status, errors.Wrap(err, "check banned")
		}

		err = cache.SetJSON(statusKey, status, 600)
		if err != nil {
			return status, errors.Wrap(err, "redis set")
		}
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

	status, err := GetMemberStatus(ctx, mid.elastic, id)
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
		rest.ResponseMessage(ctx, http.StatusForbidden, "Suspended")
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

func getMemberData(es *elastic.Client, id int) (status MemberData, err error) {
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
		return status, errors.Wrap(err, "memebr not found")
	}

	err = json.Unmarshal(searchResult.Hits.Hits[0].Source, &status)
	if err != nil {
		return status, errors.Wrap(err, "unmarshal")
	}

	return status, nil
}
