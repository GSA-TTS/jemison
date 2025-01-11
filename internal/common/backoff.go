package common

import (
	"math/rand/v2"
	"sync"
	"time"

	"go.uber.org/zap"
)

const EXPBACK = 1.03

func BackoffLoop(host string, politeSleep int64, lastHitMap *sync.Map, lastBackoffMap *sync.Map) {
	for {
		// Look at the timing map.
		lastHitTime, _ := lastHitMap.Load(host)
		// If we're in the map, and we're within 2s, we should keep checking after a backoff
		politeDuration := time.Duration(politeSleep) * time.Second

		lht, ok := lastHitTime.(time.Time)
		if !ok {
			zap.L().Error("could not cast time.Time")
		}

		if ok && (time.Since(lht) < politeDuration) {
			// There will be a last backoff time.
			last, _ := lastBackoffMap.Load(host)

			lv, ok := last.(int64)
			if !ok {
				zap.L().Error("could not cast int64")
			}

			//nolint:gosec
			newBackoffTime := float64(politeSleep)/10*rand.Float64() + float64(lv)*EXPBACK
			time.Sleep(time.Duration(newBackoffTime) * time.Second)

			continue
		}

		// We're not in the map, or it is more than <polite> milliseconds!
		// IT IS OUR TURN.
		// Reset the times and get out of here.
		zap.L().Debug("freedom: left the backoff loop",
			zap.String("host", host))
		lastBackoffMap.Store(host, politeSleep)
		lastHitMap.Store(host, time.Now())

		break
	}
}
