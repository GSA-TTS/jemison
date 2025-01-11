package main

import (
	"log"
	"net/http"

	common "github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var ThisServiceName = "validate"

var RecentlyVisitedCache *cache.Cache

func main() {
	env.InitGlobalEnv(ThisServiceName)
	InitializeQueues()

	engine := common.InitializeAPI()

	log.Println("environment initialized")

	// Init a cache for the workers
	// service, _ := env.Env.GetUserService("validate")

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	//nolint:gosec
	err := http.ListenAndServe(":"+env.Env.Port, engine)
	if err != nil {
		zap.Error(err)
	}
}
