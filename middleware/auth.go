package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/forkyid/go-utils/aes"
	"github.com/forkyid/go-utils/cache"
	"github.com/forkyid/go-utils/jwt"
	"github.com/forkyid/go-utils/rest"
	"github.com/go-redis/redis"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type MemberDataKey struct {
	ID       string `cache:"key"`
	DeviceID string `cache:"optional" json:"device_id"`
}

type MemberData struct {
	IsBanned   bool       `json:"is_banned,omitempty"`
	SuspendEnd *time.Time `json:"suspend_end,omitempty"`
	FCMToken   string     `json:"fcm_token,omitempty"`
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
		query := elastic.NewMatchQuery("id", aes.Encrypt(id))
		searchResult, err := mid.elastic.Search().
			Index("users").
			Type("_doc").
			Query(query).
			Do(context.Background())

		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).
				Log("auth: ES: " + err.Error())
			ctx.Abort()
			return
		}

		if searchResult == nil || searchResult.TotalHits() == 0 {
			rest.ResponseMessage(ctx, http.StatusNotFound)
			ctx.Abort()
			return
		}

		err = json.Unmarshal(searchResult.Hits.Hits[0].Source, &status)
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).
				Log("auth: unmarshal: " + err.Error())
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

	if status.SuspendEnd.After(time.Now()) {
		rest.ResponseMessage(ctx, http.StatusForbidden, "Suspended")
		ctx.Abort()
		return
	}

	ctx.Next()
}
