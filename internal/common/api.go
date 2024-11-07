package common

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
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

// FIXME Switch to a concurrency-safe map library...
type StatsMap = sync.Map

type Stats interface {
	Set(string, int64)
	Increment(string)
	Get(string) int64
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
		c.Request.Close()
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
	return v.(*BaseStats)
}

// extract     | fatal error: concurrent map writes
func (e *BaseStats) Set(key string, val int64) {
	e.stats.Store(key, val)
}

func (e *BaseStats) HasKey(key string) bool {
	_, ok := e.stats.Load(key)
	return ok
}

func (e *BaseStats) Get(key string) int64 {
	v, _ := e.stats.Load(key)
	return v.(int64)
}

func (e *BaseStats) GetAll() map[string]int64 {
	copy := make(map[string]int64, 0)
	e.stats.Range(func(key any, v any) bool {
		copy[key.(string)] = v.(int64)
		return true
	})
	return copy
}

func (e *BaseStats) Increment(key string) {
	if val, ok := e.stats.Load(key); ok {
		e.Set(key, val.(int64)+1)
	} else {
		e.Set(key, 1)
	}
}

func (e *BaseStats) Sum(key string, v int64) {
	if val, ok := e.stats.Load(key); ok {
		e.Set(key, val.(int64)+v)
	} else {
		e.Set(key, v)
	}
}
