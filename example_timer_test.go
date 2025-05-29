package timer_test

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jnpr-pranav/go-timer"
)

// ExampleTimer demonstrates the basic usage of the Timer.
func ExampleTimer() {
	// Create a new timer
	timer := timer.NewTimer()

	// Record the start time
	start := time.Now()
	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	// Update the timer with the duration of the work
	err := timer.Update(start)
	if err != nil {
		fmt.Printf("Error updating timer: %v\\n", err)
		return
	}

	// Print the timing statistics
	// fmt.Println(timer.String()) // Uncomment this line to see the timer's string representation
	fmt.Println(timer.Count())

	// Record another instance of work
	start = time.Now()
	time.Sleep(200 * time.Millisecond)
	err = timer.Update(start)
	if err != nil {
		fmt.Printf("Error updating timer: %v\\n", err)
	}

	// Print the timing statistics again
	// fmt.Println(timer.String()) // Uncomment this line to see the timer's string representation
	fmt.Println(timer.Count())

	// Reset the timer
	timer.Reset()
	fmt.Println(timer.Count())

	// Output:
	// 1
	// 2
	// 0
}

// ExampleTimer_concurrentUsage demonstrates using the Timer package within
// a concurrent context, where multiple goroutines update the timer.
func ExampleTimer_concurrentUsage() {
	// Create a new timer
	timer := timer.NewTimer()

	// Number of concurrent updates
	numUpdates := 100
	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numUpdates)

	// Perform concurrent updates
	for range numUpdates {
		go func() {
			defer wg.Done()
			start := time.Now()
			// Simulate some work with random duration
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			err := timer.Update(start)
			if err != nil {
				// In a real application, handle this error appropriately
				fmt.Printf("Error updating timer: %v\\n", err)
			}
		}()
	}

	// Wait for all updates to complete
	wg.Wait()

	// Print the final timing statistics
	// fmt.Println(timer.String()) // Uncomment this line to see the timer's string representation
	fmt.Println(timer.Count())

	// Output:
	// 100
}
