package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/forkyid/go-utils/aes"
	"github.com/forkyid/go-utils/cache"
	"github.com/forkyid/go-utils/jwt"
	"github.com/forkyid/go-utils/rest"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type Status struct {
	ID       string `cache:"key"`
	DeviceID string `cache:"optional" json:"device_id"`
}

type StatusData struct {
	Banned    bool
	Suspended bool
	LoggedIn  bool
}

type Authorization struct {
	elastic *elastic.Client
}

func NewAuthorization(
	elastic *elastic.Client,
) *Repository {
	return &Repository{
		elastic: elastic,
	}
}

type Authorizer interface {
	Auth(ctx *gin.Context)
}

func (auth *Authorization) Auth(ctx *gin.Context) {
	id, err := jwt.ExtractID(ctx.GetHeader("Authorization"))
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	statusKey := cache.ExternalKey("global", Status{
		ID:       id,
		DeviceID: ctx.GetHeader("X-Unique-ID"),
	})

	status := StatusData{}
	exist, err := cache.IsCacheExists(statusKey)
	if err != nil {
		rest.ResponseMessage(ctx, http.StatusInternalServerError).Log("auth: redis: check exist: " + err.Error())
		ctx.Abort()
		return
	}
	if !exist {
		query := elastic.NewMatchQuery("id", aes.Encrypt(id))
		searchResult, err := esClient.Search().
			Index("users").
			Type("_doc").
			Query(query).
			Do(context.Background())
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).Log("auth: ES: " + err.Error())
			ctx.Abort()
			return
		}

		if searchResult == nil || searchResult.TotalHits() == 0 {
			rest.ResponseError(ctx, http.StatusNotFound)
			ctx.Abort()
			return
		}

		err = json.Unmarshal(searchResult.Hits.Hits[0].Source, &status)
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).Log("auth: unmarshal: " + err.Error())
			ctx.Abort()
			return
		}

		err = cache.SetJSON(statusKey, status, 600)
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).Log("auth: cache set: " + err.Error())
			ctx.Abort()
			return
		}
	} else {
		err = cache.GetUnmarshal(statusKey, &status, 600)
		if err != nil {
			rest.ResponseMessage(ctx, http.StatusInternalServerError).Log("auth: get unmarshal: " + err.Error())
			ctx.Abort()
			return
		}
	}

	if status.Banned {
		rest.ResponseMessage(ctx, http.StatusForbidden, "Banned")
		ctx.Abort()
		return
	}

	if status.Suspended {
		rest.ResponseMessage(ctx, http.StatusForbidden, "Suspended")
		ctx.Abort()
		return
	}

	if !status.LoggedIn {
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	ctx.Next()
}
