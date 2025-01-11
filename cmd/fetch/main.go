package main

import (
	"log"
	"net/http"
	"time"

	common "github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/postgres"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

var PoliteSleep int64

var ThisServiceName = "fetch"

var RetryClient *http.Client

var Gateway *HostGateway

var JDB *postgres.JemisonDB

var ChQSHP = make(chan queueing.QSHP)

var Workers *river.Workers

var MaxFilesize int64

const BytesPerKb = 1024

const KbPerMb = 1024

func main() {
	env.InitGlobalEnv(ThisServiceName)
	InitializeQueues()

	JDB = postgres.NewJemisonDB()

	engine := common.InitializeAPI()
	ExtendAPI(engine)

	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryMax = 10
	RetryClient = retryableClient.StandardClient()

	log.Println("environment initialized")

	// Init a cache for the workers
	service, _ := env.Env.GetUserService(ThisServiceName)

	// Pre-compute/lookup the sleep duration for backoff
	PoliteSleep = service.GetParamInt64("polite_sleep")
	// 1024KB * 1024B => MB
	MaxFilesize = service.GetParamInt64("max_filesize_mb") * BytesPerKb * KbPerMb

	loggerLevel := service.GetParamString("debug_level")
	if loggerLevel != "debug" {
		retryableClient.Logger = nil
	}

	Gateway = NewHostGateway(time.Duration(PoliteSleep) * time.Second)

	go InfoFetchCount()

	go queueing.Enqueue(ChQSHP)
	go queueing.ClearCompletedPeriodically()

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	//nolint:gosec
	err := http.ListenAndServe(":"+env.Env.Port, engine)
	if err != nil {
		zap.Error(err)
	}
}
