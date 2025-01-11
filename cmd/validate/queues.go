package main

import (
	"context"
	"fmt"
	"log"
	"os"

	common "github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

type FetchWorker struct {
	river.WorkerDefaults[common.FetchArgs]
}

//nolint:lll
func initX[T river.Worker[U], U river.JobArgs](serviceName string, queueName string, workerStruct T) *river.Client[pgx.Tx] {
	queueing.InitializeRiverQueues()

	ctx, pool, workers := common.CommonQueueInit()

	// Essentially adds a worker "type" to the work engine.
	river.AddWorker(workers, workerStruct)

	// Grab the number of workers from the config.
	theService, err := env.Env.GetUserService(serviceName)
	if err != nil {
		zap.L().Error("could not fetch service config",
			zap.String("service_name", serviceName))
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	theClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			queueName: {MaxWorkers: int(theService.GetParamInt64("workers"))},
		},
		Workers: workers,
	})
	if err != nil {
		zap.L().Error("could not establish worker pool",
			zap.String("service_name", serviceName),
			zap.String("queue_name", queueName),
			zap.String("error", fmt.Sprintln(err)))
		log.Println(err)
		os.Exit(1)
	}

	// Start the work clients
	if err := theClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.",
			zap.String("queue_name", queueName))
		os.Exit(1)
	}

	return theClient
}

type ValidateFetchWorker struct {
	river.WorkerDefaults[common.ValidateFetchArgs]
}

//nolint:revive
func (w ValidateFetchWorker) Work(_ context.Context, job *river.Job[common.ValidateFetchArgs]) error {
	zap.L().Info("VALIDATE IS RUNNING AND DOING NOTHING")

	return nil
}

func InitializeQueues() {
	queueing.InitializeRiverQueues()
	initX(ThisServiceName, common.ValidateFetchQueue, ValidateFetchWorker{})
}
