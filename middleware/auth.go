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

type MemberDataKey struct {
	ID       string `cache:"key"`
	DeviceID string `cache:"optional" json:"device_id"`
}

type MemberData struct {
	IsBanned   bool                   `json:"is_banned,omitempty"`
	SuspendEnd *time.Time             `json:"suspend_end,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

func (mid *Middleware) Auth(ctx *gin.Context) {
	id, err := jwt.ExtractID(ctx.GetHeader("Authorization"))
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	statusKey := cache.ExternalKey("global", MemberDataKey{
		ID:       aes.Encrypt(id),
		DeviceID: ctx.GetHeader("X-Unique-ID"),
	})

	status := MemberData{}
	err = cache.GetUnmarshal(statusKey, &status, 600)
	if err != nil && err != redis.Nil {
		rest.ResponseMessage(ctx, http.StatusInternalServerError).
			Log("auth: get unmarshal: " + err.Error())
		ctx.Abort()
		return
	}

	if err == redis.Nil {
		banned, err := isBanned(ctx)
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).
				Log("auth: check banned: " + err.Error())
			ctx.Abort()
			return
		}
		if banned {
			status.IsBanned = true
			err = cache.SetJSON(statusKey, status, 600)
			if err != nil {
				rest.ResponseMessage(ctx, http.StatusInternalServerError).
					Log("auth: cache set: " + err.Error())
				ctx.Abort()
				return
			}
		}

		status, err := getMemberData(mid.elastic, id)
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).
				Log("auth: get data: " + err.Error())
			ctx.Abort()
			return
		}

		err = cache.SetJSON(statusKey, status, 600)
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).
				Log("auth: cache set: " + err.Error())
			ctx.Abort()
			return
		}
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
