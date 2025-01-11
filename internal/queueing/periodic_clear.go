package queueing

import (
	"context"
	"time"

	"github.com/GSA-TTS/jemison/internal/common"
	"go.uber.org/zap"
)

const PeriodicCleanupMinutes = 10

func ClearCompletedPeriodically() {
	_, pool, _ := common.CommonQueueInit()
	defer pool.Close()

	ticker := time.NewTicker(PeriodicCleanupMinutes * time.Minute)

	for {
		<-ticker.C

		zap.L().Warn("clearing completed queue")

		ctx := context.Background()

		_, err := pool.Exec(ctx, "DELETE FROM river_job WHERE state='completed'")
		if err != nil {
			zap.L().Error("failed to periodically delete jobs")
		}
	}
}
