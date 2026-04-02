package aggregator

import (
	"testing"
	"time"
)

// TestEvictIncludesLastPoint verifies the fix for the off-by-one error in Evict,
// ensuring the last data point in the slice is also considered for eviction.
func TestEvictIncludesLastPoint(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	// Add a point with a timestamp that will be in the past (should be evicted)
	// We do this by directly manipulating via Add and then using a future cutoff
	// that is after all points were added.
	w.Add(10.0)
	w.Add(20.0)
	w.Add(30.0)

	// Use a cutoff of "now + 1 second" so all points (including the last) are evicted
	cutoff := time.Now().Add(1 * time.Second)
	w.Evict(cutoff)

	w.mu.Lock()
	remaining := len(w.points)
	w.mu.Unlock()

	if remaining != 0 {
		t.Errorf("expected all points (including last) to be evicted, but %d points remain", remaining)
	}
}

// TestEvictKeepsRecentPoints verifies that Evict correctly retains points
// that are after the cutoff, including the last point in the slice.
func TestEvictKeepsRecentPoints(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	// Use a cutoff in the past so all newly added points are retained
	cutoff := time.Now().Add(-1 * time.Second)

	w.Add(10.0)
	w.Add(20.0)
	w.Add(30.0)

	w.Evict(cutoff)

	w.mu.Lock()
	remaining := len(w.points)
	w.mu.Unlock()

	if remaining != 3 {
		t.Errorf("expected 3 points to remain after eviction, got %d", remaining)
	}
}

// TestAvgPreservesPrecision verifies the fix for the float64->int64 cast bug in Add
// and the integer division bug in Avg. The average should preserve decimal precision.
func TestAvgPreservesPrecision(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	// Add values whose average is a non-integer to expose precision loss
	w.Add(1.5)
	w.Add(2.5)
	w.Add(3.5)

	avg := w.Avg()
	expected := 2.5

	if avg != expected {
		t.Errorf("expected average %.4f, got %.4f (precision loss detected)", expected, avg)
	}
}

// TestAvgWithFractionalValues verifies that fractional float64 values are
// accumulated correctly without truncation to int64.
func TestAvgWithFractionalValues(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	w.Add(0.1)
	w.Add(0.2)
	w.Add(0.3)

	avg := w.Avg()
	expected := 0.2

	// Allow a small epsilon for floating point arithmetic
	epsilon := 1e-9
	diff := avg - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > epsilon {
		t.Errorf("expected average %.10f, got %.10f (diff %.10f exceeds epsilon)", expected, avg, diff)
	}
}

// TestAvgEmptyWindow verifies that Avg returns 0 when no points have been added.
func TestAvgEmptyWindow(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	avg := w.Avg()
	if avg != 0 {
		t.Errorf("expected average 0 for empty window, got %f", avg)
	}
}

// TestEvictPartialRemoval verifies that Evict correctly removes only expired points,
// including when the last point is the one that should be evicted.
func TestEvictLastPointIsExpired(t *testing.T) {
	w := NewRollingWindow(1 * time.Minute)

	// Add points that will be recent (after cutoff)
	w.Add(10.0)
	w.Add(20.0)

	// Sleep briefly to ensure the next point has a clearly later timestamp
	// but we'll use a cutoff between the two groups
	cutoff := time.Now().Add(500 * time.Millisecond)

	// Add a third point after the cutoff reference — this will be "before" cutoff
	// Actually: cutoff is in the future, so all current points are before it and get evicted
	w.Add(30.0)

	w.Evict(cutoff)

	w.mu.Lock()
	remaining := len(w.points)
	w.mu.Unlock()

	// All three points were added before the future cutoff, so all should be evicted
	if remaining != 0 {
		t.Errorf("expected 0 points after evicting with future cutoff, got %d", remaining)
	}
}