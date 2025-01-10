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

var all_the_stats *AllStats

type HandlerFunType = func(ctx context.Context, input *StatsInput) (*StatsResponseBody, error)

type StatsResponseBody struct {
	Body *StatsResponse
}

func StatsHandler(stats_base string) func(c *gin.Context) {
	return func(c *gin.Context) {
		b := NewBaseStats(stats_base)
		c.IndentedJSON(http.StatusOK, gin.H{
			"stats":    b.GetAll(),
			"response": "ok",
		})
	}
}

func NewBaseStats(service string) *BaseStats {
	if all_the_stats == nil {
		all_the_stats = &AllStats{}
	}

	if _, ok := all_the_stats.services.Load(service); !ok {
		all_the_stats.services.Store(service, &BaseStats{})
	}

	v, _ := all_the_stats.services.Load(service)

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
	a_copy := make(map[string]int64, 0)

	e.stats.Range(func(key any, v any) bool {
		val, ok := v.(int64)
		if !ok {
			zap.L().Error("could not case int64")
		}

		k, ok := key.(string)
		if !ok {
			zap.L().Error("could not case string")
		}

		a_copy[k] = val

		return true
	})

	return a_copy
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
