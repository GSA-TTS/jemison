package main

import (
	"log"
	"os"

	common "github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

// GLOBAL TO THE APP
// One pool of connections for River.

// The work client, doing the work of `Entree`.
var EntreePool *pgxpool.Pool

var EntreeClient *river.Client[pgx.Tx]

type EntreeWorker struct {
	river.WorkerDefaults[common.EntreeArgs]
}

func InitializeQueues() {
	queueing.InitializeRiverQueues()

	ctx, fP, workers := common.CommonQueueInit()
	EntreePool = fP

	// Essentially adds a worker "type" to the work engine.
	river.AddWorker(workers, &EntreeWorker{})

	// Grab the number of workers from the config.
	EntreeService, err := env.Env.GetUserService(ThisServiceName)
	if err != nil {
		zap.L().Error("could not Entree service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	EntreeClient, err = river.NewClient(riverpgxv5.New(EntreePool), &river.Config{
		Queues: map[string]river.QueueConfig{
			ThisServiceName: {MaxWorkers: int(EntreeService.GetParamInt64("workers"))},
		},
		Workers: workers,
	})
	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Start the work clients
	if err := EntreeClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(1)
	}
}
