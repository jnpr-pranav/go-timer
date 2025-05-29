package timer

import (
	"math"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewTimer(t *testing.T) {
	timer := NewTimer()

	if timer.count != 0 {
		t.Errorf("Expected count to be 0, got %d", timer.count)
	}

	if timer.max != 0 {
		t.Errorf("Expected max to be 0, got %v", timer.max)
	}

	if timer.min != time.Duration(math.MaxInt64) {
		t.Errorf("Expected min to be math.MaxInt64, got %v", timer.min)
	}
}

func TestUpdate(t *testing.T) {
	timer := NewTimer()

	// Test first update
	start := time.Now().Add(-100 * time.Millisecond)
	err := timer.Update(start)
	if err != nil {
		t.Errorf("Unexpected error on first update: %v", err)
	}

	if timer.Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", timer.Count())
	}

	if timer.Min() != timer.Max() {
		t.Errorf("Expected min and max to be equal after first update, min: %v, max: %v", timer.Min(), timer.Max())
	}

	if timer.Mean() != timer.Min() {
		t.Errorf("Expected mean to equal min after first update, mean: %v, min: %v", timer.Mean(), timer.Min())
	}

	// Test second update with smaller duration
	start = time.Now().Add(-50 * time.Millisecond)
	err = timer.Update(start)
	if err != nil {
		t.Errorf("Unexpected error on second update: %v", err)
	}

	if timer.Count() != 2 {
		t.Errorf("Expected count to be 2, got %d", timer.Count())
	}

	// Min should have decreased
	if timer.Min() >= 100*time.Millisecond {
		t.Errorf("Expected min to decrease after update with smaller duration, got %v", timer.Min())
	}

	// Max should remain the same
	if timer.Max() < 100*time.Millisecond {
		t.Errorf("Expected max to remain around 100ms after update with smaller duration, got %v", timer.Max())
	}

	// Test third update with larger duration
	start = time.Now().Add(-200 * time.Millisecond)
	err = timer.Update(start)
	if err != nil {
		t.Errorf("Unexpected error on third update: %v", err)
	}

	if timer.Count() != 3 {
		t.Errorf("Expected count to be 3, got %d", timer.Count())
	}

	// Min should remain the same
	if timer.Min() >= 100*time.Millisecond {
		t.Errorf("Expected min to remain the same after update with larger duration, got %v", timer.Min())
	}

	// Max should have increased
	if timer.Max() <= 100*time.Millisecond {
		t.Errorf("Expected max to increase after update with larger duration, got %v", timer.Max())
	}
}

func TestGetterMethods(t *testing.T) {
	timer := NewTimer()

	start := time.Now().Add(-100 * time.Millisecond)
	err := timer.Update(start)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if timer.Count() != 1 {
		t.Errorf("Expected Count() to return 1, got %d", timer.Count())
	}

	if timer.Max() < 50*time.Millisecond {
		t.Errorf("Expected Max() to return approximately 100ms, got %v", timer.Max())
	}

	if timer.Min() < 50*time.Millisecond {
		t.Errorf("Expected Min() to return approximately 100ms, got %v", timer.Min())
	}

	if timer.Mean() < 50*time.Millisecond {
		t.Errorf("Expected Mean() to return approximately 100ms, got %v", timer.Mean())
	}
}

func TestReset(t *testing.T) {
	timer := NewTimer()

	// Update a few times
	err1 := timer.Update(time.Now().Add(-100 * time.Millisecond))
	err2 := timer.Update(time.Now().Add(-200 * time.Millisecond))

	if err1 != nil || err2 != nil {
		t.Errorf("Unexpected errors: %v, %v", err1, err2)
	}

	// Simulate overflow
	timer.mutex.Lock()
	timer.totalSum = math.MaxInt64
	timer.sumOverflowed = true
	timer.mutex.Unlock()

	// Verify timer has data and overflow flag
	if timer.Count() != 2 {
		t.Errorf("Expected count to be 2 before reset, got %d", timer.Count())
	}
	if !timer.SumOverflowed() {
		t.Errorf("Expected sumOverflowed to be true before reset")
	}

	// Reset the timer
	timer.Reset()

	// Verify timer is reset to initial state
	if timer.Count() != 0 {
		t.Errorf("Expected count to be 0 after reset, got %d", timer.Count())
	}

	if timer.Max() != 0 {
		t.Errorf("Expected max to be 0 after reset, got %v", timer.Max())
	}

	if timer.Min() != time.Duration(math.MaxInt64) {
		t.Errorf("Expected min to be math.MaxInt64 after reset, got %v", timer.Min())
	}

	if timer.Mean() != 0 {
		t.Errorf("Expected mean to be 0 after reset, got %v", timer.Mean())
	}
	if timer.SumOverflowed() {
		t.Errorf("Expected sumOverflowed to be false after reset")
	}
}

func TestObserve(t *testing.T) {
	t0 := NewTimer()
	// feed in 10ms, 20ms, 5ms
	t0.Observe(10 * time.Millisecond)
	t0.Observe(20 * time.Millisecond)
	t0.Observe(5 * time.Millisecond)

	if t0.Count() != 3 {
		t.Fatalf("Count = %d; want 3", t0.Count())
	}
	if got, want := t0.Min(), 5*time.Millisecond; got != want {
		t.Errorf("Min = %v; want %v", got, want)
	}
	if got, want := t0.Max(), 20*time.Millisecond; got != want {
		t.Errorf("Max = %v; want %v", got, want)
	}
	// nearest‐nanosecond mean = (35_000_000ns + 1) / 3 = 11_666_667ns
	if got, want := t0.Mean(), 11_666_667*time.Nanosecond; got != want {
		t.Errorf("Mean = %v; want %v", got, want)
	}
}

func TestString(t *testing.T) {
	timer := NewTimer()

	// Update once
	start := time.Now().Add(-100 * time.Millisecond)
	err := timer.Update(start)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	str := timer.String()

	// Check that the string contains all the expected parts
	if !strings.Contains(str, "Count: 1") {
		t.Errorf("Expected string to contain 'Count: 1', got %s", str)
	}

	if !strings.Contains(str, "Max:") {
		t.Errorf("Expected string to contain 'Max:', got %s", str)
	}

	if !strings.Contains(str, "Min:") {
		t.Errorf("Expected string to contain 'Min:', got %s", str)
	}

	if !strings.Contains(str, "Mean:") {
		t.Errorf("Expected string to contain 'Mean:', got %s", str)
	}

	if strings.Contains(str, "(sum overflowed, mean is approximate)") {
		t.Errorf("Expected string NOT to contain overflow message, got %s", str)
	}

	// Simulate overflow
	timer.mutex.Lock()
	timer.totalSum = math.MaxInt64
	timer.sumOverflowed = true
	timer.mutex.Unlock()

	strOverflow := timer.String()
	if !strings.Contains(strOverflow, "(sum overflowed, mean is approximate)") {
		t.Errorf("Expected string to contain overflow message, got %s", strOverflow)
	}
}

func TestSumOverflow(t *testing.T) {
	timer := NewTimer()

	if timer.SumOverflowed() {
		t.Errorf("Expected SumOverflowed to be false for a new timer")
	}

	// Simulate a large duration that doesn't overflow yet
	err := timer.Update(time.Now().Add(-time.Duration(math.MaxInt64 / 2)))
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if timer.SumOverflowed() {
		t.Errorf("Expected SumOverflowed to be false after one large update")
	}
	if timer.totalSum != math.MaxInt64/2 {
		t.Errorf("Expected totalSum to be math.MaxInt64/2, got %d", timer.totalSum)
	}

	// Simulate another large duration that causes overflow
	err = timer.Update(time.Now().Add(-time.Duration(math.MaxInt64/2 + 1000))) // 1000ns more to ensure overflow
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !timer.SumOverflowed() {
		t.Errorf("Expected SumOverflowed to be true after overflow")
	}
	if timer.totalSum != math.MaxInt64 {
		t.Errorf("Expected totalSum to be capped at math.MaxInt64, got %d", timer.totalSum)
	}

	// Add another small duration, sum should remain capped
	currentSum := timer.totalSum
	err = timer.Update(time.Now().Add(-time.Nanosecond))
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if timer.totalSum != currentSum {
		t.Errorf("Expected totalSum to remain capped at %d after overflow, got %d", currentSum, timer.totalSum)
	}
	if !timer.SumOverflowed() {
		t.Errorf("Expected SumOverflowed to remain true")
	}

	timer.Reset()
	if timer.SumOverflowed() {
		t.Errorf("Expected SumOverflowed to be false after reset")
	}
}

func TestUpdateWithZeroTime(t *testing.T) {
	timer := NewTimer()
	err := timer.Update(time.Time{})
	if err == nil {
		t.Errorf("Expected error when updating with zero time, got nil")
	}
	if timer.Count() != 0 {
		t.Errorf("Expected count to be 0 after zero time update, got %d", timer.Count())
	}
}

func TestUpdateWithNegativeDuration(t *testing.T) {
	timer := NewTimer()
	// time.Now() is later than start, so duration is positive
	// To simulate a negative duration effectively, we'd need to manipulate time.Now()
	// or pass a start time that is in the future.
	// The current implementation handles durNano < 0 by setting it to 0.
	// Let's test this path by providing a start time that is in the future.
	start := time.Now().Add(100 * time.Millisecond) // Start time in the future
	err := timer.Update(start)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if timer.Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", timer.Count())
	}
	// Duration should be 0
	if timer.Min() != 0 {
		t.Errorf("Expected min to be 0 for negative duration, got %v", timer.Min())
	}
	if timer.Max() != 0 {
		t.Errorf("Expected max to be 0 for negative duration, got %v", timer.Max())
	}
	if timer.Mean() != 0 {
		t.Errorf("Expected mean to be 0 for negative duration, got %v", timer.Mean())
	}
	if timer.totalSum != 0 {
		t.Errorf("Expected totalSum to be 0 for negative duration, got %v", timer.totalSum)
	}
}

func TestConcurrentUpdates(t *testing.T) {
	timer := NewTimer()
	iterations := 100
	var wg sync.WaitGroup
	var errCount int
	var mu sync.Mutex

	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			// Vary the durations to test min/max functionality
			delay := time.Duration(50+i%100) * time.Millisecond
			start := time.Now().Add(-delay)
			if err := timer.Update(start); err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if errCount > 0 {
		t.Errorf("%d errors occurred during concurrent updates", errCount)
	}

	if timer.Count() != uint64(iterations) {
		t.Errorf("Expected count to be %d after concurrent updates, got %d", iterations, timer.Count())
	}

	// Min should be approximately 50ms
	if timer.Min() > 100*time.Millisecond {
		t.Errorf("Expected min to be approximately 50ms, got %v", timer.Min())
	}

	// Max should be approximately 150ms
	if timer.Max() < 100*time.Millisecond {
		t.Errorf("Expected max to be approximately 150ms, got %v", timer.Max())
	}
}

func TestUpdateWithDifferentDurations(t *testing.T) {
	timer := NewTimer()

	// Create a range of durations to test
	durations := []time.Duration{
		10 * time.Microsecond,
		100 * time.Microsecond,
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		1 * time.Second,
	}

	for _, duration := range durations {
		start := time.Now().Add(-duration)
		err := timer.Update(start)
		if err != nil {
			t.Errorf("Unexpected error with duration %v: %v", duration, err)
		}
	}

	if timer.Count() != uint64(len(durations)) {
		t.Errorf("Expected count to be %d, got %d", len(durations), timer.Count())
	}

	if timer.Min() > 1*time.Millisecond {
		t.Errorf("Expected min to be less than 1ms, got %v", timer.Min())
	}

	if timer.Max() < 500*time.Millisecond {
		t.Errorf("Expected max to be approximately 1s, got %v", timer.Max())
	}
}

func BenchmarkTimerUpdate(b *testing.B) {
	timer := NewTimer()

	inputs := make([]time.Time, 100)
	baseTime := time.Now()
	for i := range 100 {
		inputs[i] = baseTime.Add(-time.Duration(i%100) * time.Microsecond)
	}

	idx := 0
	for b.Loop() {
		err := timer.Update(inputs[idx%100])
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		idx++
	}
}

func BenchmarkTimerConcurrentUpdate(b *testing.B) {
	timer := NewTimer()

	inputs := make([]time.Time, 100)
	baseTime := time.Now()
	for i := range 100 {
		inputs[i] = baseTime.Add(-time.Duration(i%100) * time.Microsecond)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		localInputIndex := 0

		for pb.Next() {
			currentInput := inputs[localInputIndex%100]

			// Ignore errors in benchmark
			_ = timer.Update(currentInput)

			localInputIndex++
		}
	})
}
