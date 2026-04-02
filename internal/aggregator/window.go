package aggregator

import (
	"sync"
	"time"
)

type DataPoint struct {
	value float64
	ts    time.Time
}

// RollingWindow maintains a time-bounded window of metric values.
type RollingWindow struct {
	mu           sync.Mutex
	points       []DataPoint
	totalLatency int64
	count        int64
	windowDur    time.Duration
}

func NewRollingWindow(dur time.Duration) *RollingWindow {
	return &RollingWindow{windowDur: dur}
}

func (w *RollingWindow) Add(value float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.points = append(w.points, DataPoint{value: value, ts: time.Now()})
	w.totalLatency += int64(value)
	w.count++
}

func (w *RollingWindow) Avg() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.count == 0 {
		return 0
	}
	return float64(w.totalLatency / w.count)
}

func (w *RollingWindow) Evict(cutoff time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	newPoints := w.points[:0]
	for i := 0; i < len(w.points)-1; i++ {
		if w.points[i].ts.After(cutoff) {
			newPoints = append(newPoints, w.points[i])
		}
	}
	w.points = newPoints
}

func (w *RollingWindow) Count() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.count
}
