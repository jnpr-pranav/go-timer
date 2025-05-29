// Package timer provides a concurrent-safe utility for tracking execution durations
// and calculating statistics like min, max, and mean times.
package timer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Timer tracks execution durations with thread-safe statistics collection.
// All methods are safe for concurrent use.
type Timer struct {
	mutex sync.RWMutex
	count uint64        // Number of durations observed
	max   time.Duration // Maximum observed duration
	min   time.Duration // Minimum observed duration
	// Total sum of all durations in nanoseconds (may be capped at MaxInt64)
	totalSum int64
	// Indicates if totalSum reached MaxInt64 and was capped
	sumOverflowed bool
}

// NewTimer creates a new Timer with initialized min/max values.
func NewTimer() *Timer {
	return &Timer{
		max: 0,
		min: time.Duration(math.MaxInt64),
	}
}

// Observe records a duration in the timer statistics.
// Thread-safe and can be called concurrently from multiple goroutines.
func (t *Timer) Observe(d time.Duration) {
	durNano := d.Nanoseconds()
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.count == 0 {
		t.min, t.max = d, d
	} else {
		if d < t.min {
			t.min = d
		}
		if d > t.max {
			t.max = d
		}
	}

	// cap at MaxInt64, set overflow flag if needed
	if durNano > 0 && t.totalSum > math.MaxInt64-durNano {
		t.totalSum = math.MaxInt64
		t.sumOverflowed = true
	} else if !t.sumOverflowed {
		t.totalSum += durNano
	}

	t.count++
}

// Update calculates the duration since the provided start time and records it.
// Returns an error if start is a zero time value.
// The duration is clamped to non-negative values.
func (t *Timer) Update(start time.Time) error {
	if start.IsZero() {
		return fmt.Errorf("cannot update timer with zero time value")
	}
	d := max(time.Since(start), 0)
	t.Observe(d)
	return nil
}

// Count returns the number of observations recorded.
func (t *Timer) Count() uint64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.count
}

// Max returns the maximum duration observed.
// Returns 0 if no observations have been made.
func (t *Timer) Max() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.max
}

// Min returns the minimum duration observed.
// Returns a very large value if no observations have been made.
func (t *Timer) Min() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.min
}

// meanNoLock calculates the mean duration without acquiring a lock.
// Used internally by Mean() and String() to avoid lock acquisition overhead.
// Adds half the count to achieve proper rounding rather than truncation.
// Returns 0 if no observations have been made.
func (t *Timer) meanNoLock() time.Duration {
	if t.count == 0 {
		return 0
	}
	// add half a count to round and not floor
	meanNano := (t.totalSum + int64(t.count)/2) / int64(t.count)
	return time.Duration(meanNano)
}

// Mean returns the average of all observed durations.
// Uses integer division with rounding to calculate the average.
// Returns 0 if no observations have been made.
func (t *Timer) Mean() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.meanNoLock()
}

// Reset clears all statistics and returns the timer to its initial state.
func (t *Timer) Reset() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.count = 0
	t.totalSum = 0
	t.max = 0
	t.min = time.Duration(math.MaxInt64)
	t.sumOverflowed = false // Reset the flag
}

// SumOverflowed returns true if the total sum of durations has exceeded
// math.MaxInt64 nanoseconds, causing the mean to be an underestimate.
func (t *Timer) SumOverflowed() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.sumOverflowed
}

// String returns a human-readable representation of the timer's statistics.
// Format: "Count: X, Max: Xms, Min: Xms, Mean: Xms"
// Includes an overflow indicator if applicable.
func (t *Timer) String() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	c, mx, mn, mean, overflowed := t.count, t.max, t.min, t.meanNoLock(), t.sumOverflowed

	var sb strings.Builder
	sb.Grow(150)
	sb.WriteString("Count: ")
	sb.WriteString(strconv.FormatUint(c, 10))
	sb.WriteString(", Max: ")
	sb.WriteString(mx.String())
	sb.WriteString(", Min: ")
	sb.WriteString(mn.String())
	sb.WriteString(", Mean: ")
	sb.WriteString(mean.String())
	if overflowed {
		sb.WriteString(" (sum overflowed, mean is approximate)")
	}
	return sb.String()
}
