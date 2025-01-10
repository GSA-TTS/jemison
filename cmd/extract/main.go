package main

import (
	"log"
	"net/http"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"go.uber.org/zap"
)

var ThisServiceName = "extract"

var ChQSHP = make(chan queueing.QSHP)

func main() {
	env.InitGlobalEnv(ThisServiceName)
	InitializeQueues()
	log.Println("environment initialized")

	routers := common.InitializeAPI()

	go queueing.Enqueue(ChQSHP)

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	err := http.ListenAndServe(":"+env.Env.Port, routers)
	if err != nil {
		zap.Error(err)
	}
}
