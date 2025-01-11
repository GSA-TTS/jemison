package main

import (
	"reflect"
	"sync"
)

const mutexLocked = 1

func MutexLocked(m *sync.Mutex) bool {
	state := reflect.ValueOf(m).Elem().FieldByName("state")

	return state.Int()&mutexLocked == mutexLocked
}

type PerHostLock struct {
	m         *sync.Mutex
	hostlocks map[string]*sync.Mutex
}

func NewPerHostLock() *PerHostLock {
	return &PerHostLock{
		m:         &sync.Mutex{},
		hostlocks: make(map[string]*sync.Mutex),
	}
}

func (phl *PerHostLock) Lock(host string) {
	phl.m.Lock()
	defer phl.m.Unlock()

	if _, ok := phl.hostlocks[host]; !ok {
		phl.hostlocks[host] = &sync.Mutex{}
	}

	// Lock this host
	phl.hostlocks[host].Lock()
}

func (phl *PerHostLock) Unlock(host string) {
	phl.m.Lock()
	defer phl.m.Unlock()

	phl.hostlocks[host].Unlock()
}
