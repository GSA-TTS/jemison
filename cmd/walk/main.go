package main

import (
	"log"
	"net/http"
	"time"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/queueing"
	expirable "github.com/go-pkgz/expirable-cache/v3"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var expirable_cache expirable.Cache[string, int]

var RecentlyVisitedCache *cache.Cache

var ChQSHP = make(chan queueing.QSHP)

var ThisServiceName = "walk"

func main() {
	env.InitGlobalEnv(ThisServiceName)

	InitializeQueues()

	log.Println("environment initialized")

	service, _ := env.Env.GetUserService("walk")

	engine := common.InitializeAPI()

	ttl := service.GetParamInt64("cache-ttl")
	expirable_cache = expirable.NewCache[string, int]().WithTTL(time.Duration(ttl) * time.Second)

	RecentlyVisitedCache = cache.New(
		time.Duration(service.GetParamInt64("polite_cache_default_expiration"))*time.Second,
		time.Duration(service.GetParamInt64("polite_cache_cleanup_interval"))*time.Second)

	go queueing.Enqueue(ChQSHP)

	log.Println("WALK listening on", env.Env.Port)

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	err := http.ListenAndServe(":"+env.Env.Port, engine)
	if err != nil {
		zap.L().Fatal("failed to start http server in serve")
	}
}
