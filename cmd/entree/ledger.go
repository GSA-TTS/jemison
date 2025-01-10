package main

import "sync"

type SafeLedger struct {
	mu     sync.RWMutex
	Ledger map[string]bool
}

func NewSafeLedger() *SafeLedger {
	l := &SafeLedger{
		Ledger: make(map[string]bool),
	}

	return l
}

func (l *SafeLedger) Add(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Ledger[entry] = true
}

func (l *SafeLedger) Check(entry string) bool {
	return l.Ledger[entry]
}

func (l *SafeLedger) Remove(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.Ledger, entry)
}

var HallPassLedger *SafeLedger = NewSafeLedger()
