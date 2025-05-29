package timer

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Timer struct {
	mutex sync.Mutex
	count uint64
	max   time.Duration
	min   time.Duration
	mean  time.Duration
}

func NewTimer() *Timer {
	return &Timer{
		count: 0,
		max:   0,
		min:   time.Duration(math.MaxInt64),
		mean:  0,
	}
}

func (t *Timer) Update(start time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	duration := time.Since(start)

	// Initialize min and max on the first update
	if t.count == 0 {
		t.min = duration
		t.max = duration
	} else {
		// Update min and max only if the new duration is less than min or greater than max
		if duration < t.min {
			t.min = duration
		}
		if duration > t.max {
			t.max = duration
		}
	}

	// Update the count
	t.count++

	// Update the mean using a running average formula
	t.mean += (duration - t.mean) / time.Duration(t.count)
}

func (t *Timer) Count() uint64 {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.count
}

func (t *Timer) Max() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.max
}

func (t *Timer) Min() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.min
}

func (t *Timer) Mean() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.mean
}

func (t *Timer) Reset() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.count = 0
	t.max = 0
	t.min = time.Duration(math.MaxInt64)
	t.mean = 0
}

func (t *Timer) String() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return "Count: " + fmt.Sprintf("%d", t.count) +
		", Max: " + t.max.String() +
		", Min: " + t.min.String() +
		", Mean: " + t.mean.String()
}
