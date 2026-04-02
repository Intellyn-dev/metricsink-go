package aggregator

import (
	"testing"
	"time"
)

// TestAvgWithFloatValues verifies the fix for bug 1:
// totalLatency is now float64, so fractional values are preserved
// and Avg() returns a correct float64 average instead of truncated integer division.
func TestAvgWithFloatValues(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	w.Add(1.5)
	w.Add(2.5)
	w.Add(3.0)

	// Expected average: (1.5 + 2.5 + 3.0) / 3 = 2.333...
	// Before fix: totalLatency was int64, so 1.5+2.5+3.0 would truncate to 6,
	// and 6/3 = 2.0 (wrong). After fix: 7.0/3 = 2.3333...
	got := w.Avg()
	expected := 7.0 / 3.0

	if got != expected {
		t.Errorf("Avg() = %v, want %v", got, expected)
	}
}

// TestAvgFractionalNotTruncated verifies that fractional float64 values
// are not truncated when accumulated in totalLatency.
func TestAvgFractionalNotTruncated(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	w.Add(0.1)
	w.Add(0.2)

	got := w.Avg()
	expected := 0.15

	// Allow tiny floating point epsilon
	diff := got - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > 1e-9 {
		t.Errorf("Avg() = %v, want %v (diff %v)", got, expected, diff)
	}
}

// TestEvictIncludesLastElement verifies the fix for bug 2:
// The Evict() loop previously used `i < len(w.points)-1`, skipping the last element.
// After the fix, all elements including the last are evaluated against the cutoff.
func TestEvictIncludesLastElement(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	// Add a value that is clearly in the past (should be evicted)
	pastPoint := DataPoint{value: 99.0, ts: time.Now().Add(-10 * time.Minute)}
	w.mu.Lock()
	w.points = append(w.points, pastPoint)
	w.totalLatency += pastPoint.value
	w.count++
	w.mu.Unlock()

	// The past point is the only (and thus last) element.
	// Before fix: loop ran `i < len(w.points)-1` = `i < 0`, so zero iterations,
	// meaning the last element was never evicted.
	// After fix: the loop processes all elements including the last.
	cutoff := time.Now().Add(-1 * time.Minute)
	w.Evict(cutoff)

	w.mu.Lock()
	remaining := len(w.points)
	w.mu.Unlock()

	if remaining != 0 {
		t.Errorf("after Evict, expected 0 points remaining, got %d", remaining)
	}
}

// TestEvictKeepsRecentLastElement verifies that a recent last element is NOT evicted.
func TestEvictKeepsRecentLastElement(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	// Add an old point (should be evicted)
	oldPoint := DataPoint{value: 10.0, ts: time.Now().Add(-5 * time.Minute)}
	// Add a recent point as the last element (should be kept)
	recentPoint := DataPoint{value: 20.0, ts: time.Now().Add(-10 * time.Second)}

	w.mu.Lock()
	w.points = append(w.points, oldPoint, recentPoint)
	w.totalLatency += oldPoint.value + recentPoint.value
	w.count += 2
	w.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Minute)
	w.Evict(cutoff)

	w.mu.Lock()
	remaining := len(w.points)
	var remainingVal float64
	if remaining > 0 {
		remainingVal = w.points[0].value
	}
	w.mu.Unlock()

	if remaining != 1 {
		t.Errorf("after Evict, expected 1 point remaining, got %d", remaining)
	}
	if remainingVal != 20.0 {
		t.Errorf("remaining point value = %v, want 20.0", remainingVal)
	}
}

// TestEvictAllOldPoints verifies multiple old points are all evicted,
// including the last one (regression for the off-by-one fix).
func TestEvictAllOldPoints(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	cutoff := time.Now().Add(-1 * time.Minute)

	for i := 0; i < 5; i++ {
		w.Add(float64(i))
	}

	// Manually set all timestamps to the past so they should be evicted
	w.mu.Lock()
	for i := range w.points {
		w.points[i].ts = time.Now().Add(-2 * time.Minute)
	}
	w.mu.Unlock()

	w.Evict(cutoff)

	w.mu.Lock()
	remaining := len(w.points)
	w.mu.Unlock()

	if remaining != 0 {
		t.Errorf("after Evict of all old points, expected 0 remaining, got %d", remaining)
	}
}

// TestAvgZeroCountReturnsZero verifies Avg() returns 0 when no values have been added.
func TestAvgZeroCountReturnsZero(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)
	got := w.Avg()
	if got != 0 {
		t.Errorf("Avg() on empty window = %v, want 0", got)
	}
}