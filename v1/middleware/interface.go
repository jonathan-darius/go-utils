package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type Middleware struct {
	elastic *elastic.Client
}

func NewMiddleware(
	elastic *elastic.Client,
) *Middleware {
	return &Middleware{
		elastic: elastic,
	}
}

type Middlewarer interface {
	Auth(ctx *gin.Context)
	CORS(ctx *gin.Context)
	CheckFeatureFlagStatus(key string)
	CheckWaitingStatus(ctx *gin.Context)
}
