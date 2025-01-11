package common

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Heartbeat(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func InitializeAPI() *gin.Engine {
	router := gin.Default()
	router.GET("/heartbeat", Heartbeat)

	return router
}

type StatsInput struct{}

type StatsResponse struct {
	Stats map[string]int64 `json:"stats"`
}

type StatsMap = sync.Map

type Stats interface {
	Set(key string, value int64)
	Increment(key string)
	Get(key string) int64
	GetAll() StatsMap
}

type BaseStats struct {
	stats StatsMap
}

type AllStats struct {
	services sync.Map
}

var allTheStats *AllStats

type HandlerFunType = func(ctx context.Context, input *StatsInput) (*StatsResponseBody, error)

type StatsResponseBody struct {
	Body *StatsResponse
}

func StatsHandler(statsBase string) func(c *gin.Context) {
	return func(c *gin.Context) {
		b := NewBaseStats(statsBase)
		c.IndentedJSON(http.StatusOK, gin.H{
			"stats":    b.GetAll(),
			"response": "ok",
		})
	}
}

func NewBaseStats(service string) *BaseStats {
	if allTheStats == nil {
		allTheStats = &AllStats{}
	}

	if _, ok := allTheStats.services.Load(service); !ok {
		allTheStats.services.Store(service, &BaseStats{})
	}

	v, _ := allTheStats.services.Load(service)

	bs, ok := v.(*BaseStats)
	if !ok {
		zap.L().Error("could not cast basestats")
	}

	return bs
}

func (e *BaseStats) Set(key string, val int64) {
	e.stats.Store(key, val)
}

func (e *BaseStats) HasKey(key string) bool {
	_, ok := e.stats.Load(key)

	return ok
}

func (e *BaseStats) Get(key string) int64 {
	val, _ := e.stats.Load(key)

	v, ok := val.(int64)
	if !ok {
		zap.L().Error("could not case int64")
	}

	return v
}

func (e *BaseStats) GetAll() map[string]int64 {
	aCopy := make(map[string]int64, 0)

	e.stats.Range(func(key any, v any) bool {
		val, ok := v.(int64)
		if !ok {
			zap.L().Error("could not case int64")
		}

		k, ok := key.(string)
		if !ok {
			zap.L().Error("could not case string")
		}

		aCopy[k] = val

		return true
	})

	return aCopy
}

func (e *BaseStats) Increment(key string) {
	if val, ok := e.stats.Load(key); ok {
		v, ok := val.(int64)
		if !ok {
			zap.L().Error("could not cast int64")
		}

		e.Set(key, v+1)
	} else {
		e.Set(key, 1)
	}
}

func (e *BaseStats) Sum(key string, incr int64) {
	if val, ok := e.stats.Load(key); ok {
		v, ok := val.(int64)
		if !ok {
			zap.L().Error("cannot cast int64")
		}

		e.Set(key, v+incr)
	} else {
		e.Set(key, incr)
	}
}
