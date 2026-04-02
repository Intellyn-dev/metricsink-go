package aggregator

import (
	"sync"
	"time"
)

type Aggregator struct {
	mu      sync.RWMutex
	windows map[string]*RollingWindow
	dur     time.Duration
}

func NewAggregator(windowDur time.Duration) *Aggregator {
	a := &Aggregator{
		windows: make(map[string]*RollingWindow),
		dur:     windowDur,
	}
	go a.evictLoop()
	return a
}

func (a *Aggregator) key(service, metric string) string {
	return service + ":" + metric
}

func (a *Aggregator) Add(service, metric string, value float64) {
	a.mu.Lock()
	k := a.key(service, metric)
	if _, ok := a.windows[k]; !ok {
		a.windows[k] = NewRollingWindow(a.dur)
	}
	w := a.windows[k]
	a.mu.Unlock()
	w.Add(value)
}

func (a *Aggregator) GetAvg(service, metric string) float64 {
	a.mu.RLock()
	w, ok := a.windows[a.key(service, metric)]
	a.mu.RUnlock()
	if !ok {
		return 0
	}
	return w.Avg()
}

func (a *Aggregator) evictLoop() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		cutoff := time.Now().Add(-a.dur)
		a.mu.RLock()
		for _, w := range a.windows {
			w.Evict(cutoff)
		}
		a.mu.RUnlock()
	}
}
