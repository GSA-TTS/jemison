package main

import (
	"log"
	"net/http"
	"time"

	common "github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var RecentlyVisitedCache *cache.Cache
var polite_sleep int64
var ThisServiceName = "fetch"

var RetryClient *http.Client

func main() {
	env.InitGlobalEnv(ThisServiceName)
	InitializeQueues()

	engine := common.InitializeAPI()
	ExtendApi(engine)

	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryMax = 10
	//retryableClient.Logger = zap.L()
	RetryClient = retryableClient.StandardClient()

	log.Println("environment initialized")

	// Init a cache for the workers
	service, _ := env.Env.GetUserService(ThisServiceName)

	// Pre-compute/lookup the sleep duration for backoff
	polite_sleep = service.GetParamInt64("polite_sleep")

	RecentlyVisitedCache = cache.New(
		time.Duration(service.GetParamInt64("polite_cache_default_expiration"))*time.Second,
		time.Duration(service.GetParamInt64("polite_cache_cleanup_interval"))*time.Second)

	go InfoFetchCount()

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)

}
