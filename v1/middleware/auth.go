package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/forkyid/go-utils/v1/aes"
	"github.com/forkyid/go-utils/v1/cache"
	"github.com/forkyid/go-utils/v1/jwt"
	"github.com/forkyid/go-utils/v1/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
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
	isAlive := cache.IsCacheConnected()
	if !isAlive {
		log.Printf("redis connect failed: %s\n", os.Getenv("REDIS_HOST"))
	}

	statusKey := cache.ExternalKey("global", MemberStatusKey{
		ID: aes.Encrypt(memberID),
	})

	if isAlive {
		err = cache.GetUnmarshal(statusKey, &status)
		if err == nil {
			if status.SuspendEnd != nil && status.SuspendEnd.After(time.Now().Add(10*time.Minute)) {
				suspendEnd := time.Until(*status.SuspendEnd)
				cache.SetExpire(statusKey, int(suspendEnd.Seconds()))
			} else {
				cache.SetExpire(statusKey, 600)
			}
			return
		}
		if err != redis.Nil {
			log.Println("redis get unmarshal: " + err.Error())
		}
	}

	status.IsBanned, err = isBanned(ctx)
	if err != nil {
		err = errors.Wrap(err, "check banned")
		return
	}

	if isAlive {
		err = cache.SetJSON(statusKey, status, 600)
		if err != nil {
			log.Println("redis set: " + err.Error())
		}
	}

	return
}

func (mid *Middleware) Auth(ctx *gin.Context) {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		rest.ResponseMessage(ctx, http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	id, err := jwt.ExtractID(auth)
	if err != nil {
		log.Println("extract id: " + err.Error())
		rest.ResponseError(ctx, http.StatusUnauthorized, map[string]string{
			"access_token": "expired",
		})
		ctx.Abort()
		return
	}

	status, err := GetStatus(ctx, mid.elastic, id)
	if err != nil {
		log.Println("get status: " + err.Error())
		rest.ResponseMessage(ctx, http.StatusInternalServerError)
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

// CheckWaitingStatus params
//	@ctx: *gin.Context
func (m *Middleware) CheckWaitingStatus(ctx *gin.Context) {
	if err := m.elastic.WaitForYellowStatus("1s"); err != nil {
		log.Println("ERROR: wait for yellow status: " + err.Error())
		return
	}

	result, err := m.elastic.Get().
		Index("waiting-list").
		Id("status").
		Do(ctx)
	if err != nil {
		log.Println("ERROR: get waiting list status: " + err.Error())
		return
	}

	resultStruct := map[string]bool{}

	if !result.Found {
		log.Println("ERROR: waiting list status not found: " + err.Error())
		return
	}

	json.Unmarshal(result.Source, &resultStruct)
	isWait := resultStruct["status"]

	if isWait {
		rest.ResponseMessage(ctx, http.StatusServiceUnavailable, "Service Unavailable")
		ctx.Abort()
	}
}
