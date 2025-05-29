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

	if timer.mean != 0 {
		t.Errorf("Expected mean to be 0, got %v", timer.mean)
	}
}

func TestUpdate(t *testing.T) {
	timer := NewTimer()

	// Test first update
	start := time.Now().Add(-100 * time.Millisecond)
	timer.Update(start)

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
	timer.Update(start)

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
	timer.Update(start)

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
	timer.Update(start)

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
	timer.Update(time.Now().Add(-100 * time.Millisecond))
	timer.Update(time.Now().Add(-200 * time.Millisecond))

	// Verify timer has data
	if timer.Count() != 2 {
		t.Errorf("Expected count to be 2 before reset, got %d", timer.Count())
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
}

func TestString(t *testing.T) {
	timer := NewTimer()

	// Update once
	start := time.Now().Add(-100 * time.Millisecond)
	timer.Update(start)

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
}

func TestConcurrentUpdates(t *testing.T) {
	timer := NewTimer()
	iterations := 100
	var wg sync.WaitGroup

	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			// Vary the durations to test min/max functionality
			delay := time.Duration(50+i%100) * time.Millisecond
			start := time.Now().Add(-delay)
			timer.Update(start)
		}(i)
	}

	wg.Wait()

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
		timer.Update(start)
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
		timer.Update(inputs[idx%100])
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

			timer.Update(currentInput)

			localInputIndex++
		}
	})
}
