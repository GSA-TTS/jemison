package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/GSA-TTS/jemison/config"
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

const RoundRobin = "round_robin"

const OnePerDomain = "one_per_domain"

const Simple = "simple"

// The work client, doing the work of `fetch`.
var FetchPool *pgxpool.Pool

var FetchClient *river.Client[pgx.Tx]

var FetchQueues map[string]river.QueueConfig

var RoundRobinWorkerPool atomic.Int64

var RoundRobinSize int64

var QueueingModel string

type FetchWorker struct {
	river.WorkerDefaults[common.FetchArgs]
}

func oneQueuePerHost(workers *river.Workers, workerCount int64) {
	fetchService, err := env.Env.GetUserService(ThisServiceName)
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	MainQueue := make(map[string]river.QueueConfig)
	MainQueue[ThisServiceName] = river.QueueConfig{MaxWorkers: int(workerCount)}

	FetchQueues = make(map[string]river.QueueConfig)

	for _, host := range config.GetListOfHosts(env.Env.AllowedHosts) {
		asciiHost := stripHostToASCII(host)
		asciiQueueName := fmt.Sprintf("fetch-%s", asciiHost)
		zap.L().Info("setting up queue", zap.String("queue_name", asciiQueueName))

		FetchQueues[asciiQueueName] = river.QueueConfig{
			MaxWorkers: int(workerCount),
		}
	}

	// Start the 'main' queue
	mainFetchClient, err := river.NewClient(riverpgxv5.New(FetchPool), &river.Config{
		Queues:  MainQueue,
		Workers: workers,
	})
	if err != nil {
		zap.L().Error("could not establish main fetch client")
		log.Println(err)
		os.Exit(1)
	}

	hostsFetchClient, err := river.NewClient(riverpgxv5.New(FetchPool), &river.Config{
		Queues:            FetchQueues,
		Workers:           workers,
		FetchCooldown:     time.Duration(fetchService.GetParamInt64("fetch_cooldown_ms")) * time.Millisecond,
		FetchPollInterval: time.Duration(fetchService.GetParamInt64("fetch_poll_interval_ms")) * time.Millisecond,
	})
	if err != nil {
		zap.Error(err)
		zap.L().Fatal("could not establish hosts fetch client")
	}

	// Start the work clients
	ctx := context.Background()
	if err := mainFetchClient.Start(ctx); err != nil {
		zap.Error(err)
		zap.L().Fatal("could not launch main fetch client.")
	}

	ctx = context.Background()
	if err := hostsFetchClient.Start(ctx); err != nil {
		zap.Error(err)
		zap.L().Fatal("could not launch hosts")
	}
}

func roundRobinQueues(workers *river.Workers, workerCount int64) {
	// I'm going to create a round-robin pool of workers.
	// That is, if workerCount is 10, I'm going to create
	// 10x10 workers, in 10x queues. This way, I can round-robin
	// jobs to the queues. This should improve throughput significantly.
	// May have to have more configs, so this is better able to be controlled.
	RoundRobinSize = workerCount

	// Start the 'main' queue
	fetchClient, err := river.NewClient(riverpgxv5.New(FetchPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			ThisServiceName: {MaxWorkers: int(workerCount)},
		},
		Workers: workers,
	})
	if err != nil {
		zap.Error(err)
		zap.L().Fatal("could not establish hosts fetch client")
	}

	FetchClient = fetchClient

	// Start the work clients
	if err := fetchClient.Start(context.Background()); err != nil {
		zap.Error(err)
		zap.L().Fatal("workers are not the means of production")
	}

	// start the round-robin queues
	for n := range workerCount {
		queueName := fmt.Sprintf("%s-%d", ThisServiceName, n)

		fetchClient, err = river.NewClient(riverpgxv5.New(FetchPool), &river.Config{
			Queues: map[string]river.QueueConfig{
				queueName: {MaxWorkers: int(workerCount)},
			},
			// FetchCooldown: time.Duration(n*50) * time.Millisecond,
			Workers: workers,
		})
		if err != nil {
			zap.Error(err)
			zap.L().Fatal("could not establish worker pool")
		}
		// Start the work clients
		if err := fetchClient.Start(context.Background()); err != nil {
			zap.Error(err)
			zap.L().Fatal("workers are not the means of production")
		}
	}
}

func simpleQueue(workers *river.Workers, workerCount int64) {
	MainQueue := make(map[string]river.QueueConfig)
	MainQueue[ThisServiceName] = river.QueueConfig{MaxWorkers: int(workerCount)}

	// Start the 'main' queue
	mainFetchClient, err := river.NewClient(riverpgxv5.New(FetchPool), &river.Config{
		Queues:  MainQueue,
		Workers: workers,
	})
	if err != nil {
		zap.Error(err)
		zap.L().Fatal("could not establish hosts fetch client")
	}

	// Start the work clients
	ctx := context.Background()
	if err := mainFetchClient.Start(ctx); err != nil {
		zap.L().Error("could not launch main fetch client.", zap.String("err", err.Error()))
		zap.L().Fatal("exiting")
	}
}

func InitializeQueues() {
	queueing.InitializeRiverQueues()

	_, fP, workers := common.CommonQueueInit()
	FetchPool = fP

	river.AddWorker(workers, &FetchWorker{})

	// Grab the number of workers from the config.
	fetchService, err := env.Env.GetUserService(ThisServiceName)
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	workerCount := fetchService.GetParamInt64("workers")
	queueModel := fetchService.GetParamString("queue_model")

	switch queueModel {
	case RoundRobin:
		QueueingModel = RoundRobin

		roundRobinQueues(workers, workerCount)
	case OnePerDomain:
		QueueingModel = OnePerDomain

		oneQueuePerHost(workers, workerCount)
	case Simple:
		QueueingModel = Simple

		simpleQueue(workers, workerCount)
	default:
		zap.L().Warn("falling through to default simple queueing model")

		QueueingModel = Simple

		simpleQueue(workers, workerCount)
	}
}
