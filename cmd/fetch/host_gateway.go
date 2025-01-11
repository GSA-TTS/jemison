package main

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// https://medium.com/@luanrubensf/concurrent-map-access-in-go-a6a733c5ffd1
type HostGateway struct {
	politeSleep time.Duration
	last        map[string]time.Time
	m           *sync.RWMutex
}

func NewHostGateway(politeSleep time.Duration) *HostGateway {
	return &HostGateway{
		politeSleep: politeSleep,
		last:        make(map[string]time.Time),
		m:           &sync.RWMutex{},
	}
}

// When I see a host, I should check if it has a non-zero counter.
func (hsm *HostGateway) GoodToGo(host string) bool {
	hsm.m.Lock()
	defer hsm.m.Unlock()

	// Lets see if they're in the map
	lastHit, ok := hsm.last[host]

	if ok {
		// We have seen this host before
		timeSinceLastHit := time.Since(lastHit)
		// If it has been more than polite sleep since the last hit,
		// then the caller is good to go.
		isGoodToGo := timeSinceLastHit > hsm.politeSleep

		if isGoodToGo {
			// We're good to go, so we'll update the map.
			hsm.last[host] = time.Now()
			zap.L().Debug("gateway host is good to go",
				zap.Int64("timeSinceLastHit", int64(timeSinceLastHit)),
				zap.Int64("politeSleep", int64(hsm.politeSleep)),
			)
		}
		// If we were good to go, we updated the map, and should let things continue.
		// Otherwise, return false and this will be requeued.
		return isGoodToGo
	}

	/* not OK */
	// We have not seen this host before
	// Therefore, add them to the map, and they're good to go.
	zap.L().Debug("gateway: host never seen before")

	hsm.last[host] = time.Now()

	return true
}

func (hsm *HostGateway) HostExists(host string) bool {
	hsm.m.RLock()
	defer hsm.m.RUnlock()
	_, ok := hsm.last[host]

	return ok
}

func (hsm *HostGateway) TimeRemaining(host string) time.Duration {
	// This is a read operation.
	hsm.m.RLock()
	defer hsm.m.RUnlock()

	lastHit, ok := hsm.last[host]

	// We must be in the map. But, we should check.
	if ok {
		// If we're in the map, there must be some time until the
		// next possible access to the host.
		until := time.Until(lastHit.Add(hsm.politeSleep))

		zap.L().Debug("gateway time remaining",
			zap.Float64("seconds", float64(until)/float64(time.Second)))

		if until > 0 {
			return until
		}

		return time.Duration(0 * time.Millisecond)
	}

	// If someone asks for a host that is not in the map, we'll tell them
	// there are 0 milliseconds until the host is ready.
	return time.Duration(0 * time.Millisecond)
}
