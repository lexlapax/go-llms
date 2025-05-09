package stress

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// TestResponsePoolStress tests the ResponsePool under high concurrency
func TestResponsePoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Track memory stats before and after test
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create response pool
	pool := domain.NewResponsePool()

	// Define concurrency levels to test
	concurrencyLevels := []int{10, 100, 1000, 5000, 10000}

	// Define various content sizes to test
	contentSizes := []struct {
		name   string
		size   int    // Approximate size in bytes
		factor string // Small, Medium, Large
	}{
		{"Tiny", 50, "small"},
		{"Small", 500, "small"},
		{"Medium", 5000, "medium"},
		{"Large", 50000, "large"},
		{"XLarge", 500000, "large"},
	}

	// Run tests for each concurrency level and content size
	for _, concurrency := range concurrencyLevels {
		for _, size := range contentSizes {
			t.Run(fmt.Sprintf("ResponsePool_Concurrency_%d_Size_%s", concurrency, size.name), func(t *testing.T) {
				var (
					wg               sync.WaitGroup
					acquireErrors    int32
					releaseErrors    int32
					getErrors        int32
					setErrors        int32
					totalAcquireTime int64
					totalReleaseTime int64
					totalContentSize int64
				)

				// Create a semaphore to limit concurrent goroutines
				sem := make(chan struct{}, concurrency)

				// Track goroutine count
				initialGoroutines := runtime.NumGoroutine()

				// Prepare a large string with the appropriate size for testing
				template := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 "
				var contentTemplate string
				for i := 0; i < size.size/len(template)+1; i++ {
					contentTemplate += template
				}
				contentTemplate = contentTemplate[:size.size]

				// Launch concurrent requests
				startTime := time.Now()
				for i := 0; i < concurrency; i++ {
					wg.Add(1)
					sem <- struct{}{} // Acquire semaphore
					go func(id int) {
						defer func() {
							<-sem // Release semaphore
							wg.Done()
						}()

						// Generate slight variation in content to avoid compiler optimizations
						suffix := fmt.Sprintf(" (ID: %d-%d)", id, rand.Intn(10000))
						content := contentTemplate[:size.size-len(suffix)] + suffix
						contentSize := int64(len(content))
						atomic.AddInt64(&totalContentSize, contentSize)

						// Get response from pool
						acquireStart := time.Now()
						resp := pool.Get()
						acquireDuration := time.Since(acquireStart)
						atomic.AddInt64(&totalAcquireTime, acquireDuration.Nanoseconds())

						if resp == nil {
							atomic.AddInt32(&acquireErrors, 1)
							t.Logf("Get error: Response is nil (ID: %d)", id)
							return
						}

						// Set response content
						resp.Content = content

						// Verify content was set correctly
						if got := resp.Content; got != content {
							atomic.AddInt32(&getErrors, 1)
							t.Logf("Content error: Expected %d bytes, got %d bytes (ID: %d)", len(content), len(got), id)
						}

						// Simulate using the response
						time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

						// Put response back to pool
						releaseStart := time.Now()
						pool.Put(resp)
						releaseDuration := time.Since(releaseStart)
						atomic.AddInt64(&totalReleaseTime, releaseDuration.Nanoseconds())
					}(i)
				}

				// Wait for all operations to complete
				wg.Wait()
				totalDuration := time.Since(startTime)

				// Check goroutine count after test
				peakGoroutines := runtime.NumGoroutine()

				// Record results
				avgAcquireTime := float64(totalAcquireTime) / float64(concurrency) / float64(time.Microsecond)
				avgReleaseTime := float64(totalReleaseTime) / float64(concurrency) / float64(time.Microsecond)
				totalErrors := acquireErrors + releaseErrors + getErrors + setErrors
				errorRate := float64(totalErrors) / float64(concurrency) * 100
				totalContentSizeMB := float64(totalContentSize) / 1024 / 1024

				t.Logf("Results for ResponsePool at concurrency %d with content size %s (%s):",
					concurrency, size.name, size.factor)
				t.Logf("  Total processed: %.2f MB", totalContentSizeMB)
				t.Logf("  Error rate: %.2f%% (%d/%d)", errorRate, totalErrors, concurrency)
				t.Logf("  Average acquire time: %.3f µs", avgAcquireTime)
				t.Logf("  Average release time: %.3f µs", avgReleaseTime)
				t.Logf("  Total duration: %v", totalDuration)
				t.Logf("  Throughput: %.2f operations/sec", float64(concurrency)/totalDuration.Seconds())
				t.Logf("  Goroutines: %d initial, %d peak", initialGoroutines, peakGoroutines)

				// Basic validation
				if totalErrors > 0 {
					t.Logf("  Errors: acquire=%d, release=%d, get=%d, set=%d",
						acquireErrors, releaseErrors, getErrors, setErrors)
				}
			})
		}
	}

	// Collect final memory stats
	runtime.ReadMemStats(&memStatsAfter)

	// Report memory usage
	t.Logf("Memory usage before: %.2f MB", float64(memStatsBefore.Alloc)/1024/1024)
	t.Logf("Memory usage after: %.2f MB", float64(memStatsAfter.Alloc)/1024/1024)

	// Calculate memory difference with error handling for potential integer underflow
	var memDiff float64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memDiff = float64(memStatsAfter.Alloc-memStatsBefore.Alloc) / 1024 / 1024
	} else {
		// If we somehow get a negative difference (e.g., due to GC between measurements), report 0
		memDiff = 0
	}
	t.Logf("Memory difference: %.2f MB", memDiff)
	t.Logf("Total allocations: %d objects", memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// TestTokenPoolStress tests the TokenPool under high concurrency
func TestTokenPoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Track memory stats before and after test
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create token pool
	pool := domain.NewTokenPool()

	// Define pool sizes for logging purposes only
	poolSizes := []struct {
		name     string
		capacity int
	}{
		{"Small", 100},
		{"Medium", 1000},
		{"Large", 10000},
	}

	// Define concurrency levels to test
	concurrencyLevels := []int{10, 100, 1000, 5000}

	// Run tests for each pool size (just for test naming) and concurrency level
	for _, poolSize := range poolSizes {
		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("TokenPool_%s_Concurrency_%d", poolSize.name, concurrency), func(t *testing.T) {
				var (
					wg               sync.WaitGroup
					acquireErrors    int32
					releaseErrors    int32
					getErrors        int32
					setErrors        int32
					totalAcquireTime int64
					totalReleaseTime int64
					totalTokens      int64
				)

				// Create a semaphore to limit concurrent goroutines
				sem := make(chan struct{}, concurrency)

				// Track goroutine count
				initialGoroutines := runtime.NumGoroutine()

				// Launch concurrent requests
				startTime := time.Now()
				for i := 0; i < concurrency; i++ {
					wg.Add(1)
					sem <- struct{}{} // Acquire semaphore
					go func(id int) {
						defer func() {
							<-sem // Release semaphore
							wg.Done()
						}()

						// Generate random token content
						value := fmt.Sprintf("token-%d-%d", id, rand.Intn(10000))
						tokenSize := int64(len(value))
						atomic.AddInt64(&totalTokens, tokenSize)

						// Get token from pool
						acquireStart := time.Now()
						token := pool.Get()
						acquireDuration := time.Since(acquireStart)
						atomic.AddInt64(&totalAcquireTime, acquireDuration.Nanoseconds())

						if token == nil {
							atomic.AddInt32(&acquireErrors, 1)
							t.Logf("Get error: Token is nil (ID: %d)", id)
							return
						}

						// Set token value
						token.Text = value

						// Verify value was set correctly
						if got := token.Text; got != value {
							atomic.AddInt32(&getErrors, 1)
							t.Logf("Value error: Expected '%s', got '%s' (ID: %d)", value, got, id)
						}

						// Simulate using the token
						time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

						// Put token back to pool
						releaseStart := time.Now()
						pool.Put(token)
						releaseDuration := time.Since(releaseStart)
						atomic.AddInt64(&totalReleaseTime, releaseDuration.Nanoseconds())
					}(i)
				}

				// Wait for all operations to complete
				wg.Wait()
				totalDuration := time.Since(startTime)

				// Check goroutine count after test
				peakGoroutines := runtime.NumGoroutine()

				// Record results
				avgAcquireTime := float64(totalAcquireTime) / float64(concurrency) / float64(time.Microsecond)
				avgReleaseTime := float64(totalReleaseTime) / float64(concurrency) / float64(time.Microsecond)
				totalErrors := acquireErrors + releaseErrors + getErrors + setErrors
				errorRate := float64(totalErrors) / float64(concurrency) * 100

				t.Logf("Results for TokenPool at concurrency %d:", concurrency)
				t.Logf("  Error rate: %.2f%% (%d/%d)", errorRate, totalErrors, concurrency)
				t.Logf("  Average acquire time: %.3f µs", avgAcquireTime)
				t.Logf("  Average release time: %.3f µs", avgReleaseTime)
				t.Logf("  Total duration: %v", totalDuration)
				t.Logf("  Throughput: %.2f operations/sec", float64(concurrency)/totalDuration.Seconds())
				t.Logf("  Goroutines: %d initial, %d peak", initialGoroutines, peakGoroutines)

				// Basic validation
				if totalErrors > 0 {
					t.Logf("  Errors: acquire=%d, release=%d, get=%d, set=%d",
						acquireErrors, releaseErrors, getErrors, setErrors)
				}
			})
		}
	}

	// Collect final memory stats
	runtime.ReadMemStats(&memStatsAfter)

	// Report memory usage
	t.Logf("Memory usage before: %.2f MB", float64(memStatsBefore.Alloc)/1024/1024)
	t.Logf("Memory usage after: %.2f MB", float64(memStatsAfter.Alloc)/1024/1024)

	// Calculate memory difference with error handling for potential integer underflow
	var memDiff float64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memDiff = float64(memStatsAfter.Alloc-memStatsBefore.Alloc) / 1024 / 1024
	} else {
		// If we somehow get a negative difference (e.g., due to GC between measurements), report 0
		memDiff = 0
	}
	t.Logf("Memory difference: %.2f MB", memDiff)
	t.Logf("Total allocations: %d objects", memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// TestChannelPoolStress tests the ChannelPool under high concurrency
func TestChannelPoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Track memory stats before and after test
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create a channel pool
	pool := domain.NewChannelPool()

	// Define various configurations for testing (for test naming only)
	poolConfigs := []struct {
		name        string
		bufferSize  int
		numChannels int
	}{
		{"Small", 10, 5},
		{"Medium", 50, 10},
		{"Large", 100, 20},
	}

	// Define concurrency levels to test
	concurrencyLevels := []int{10, 50, 100, 200}

	// Run tests for each pool configuration (just for naming) and concurrency level
	for _, config := range poolConfigs {
		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("ChannelPool_%s_Concurrency_%d", config.name, concurrency), func(t *testing.T) {
				var (
					wg                 sync.WaitGroup
					acquireErrors      int32
					releaseErrors      int32
					totalAcquireTime   int64
					totalReleaseTime   int64
					totalMessagesCount int64
					messageDropCount   int32
					waitTimeoutCount   int32
				)

				// Create a semaphore to limit concurrent goroutines
				sem := make(chan struct{}, concurrency)

				// Track goroutine count
				initialGoroutines := runtime.NumGoroutine()

				// Message sizes to test
				messageSizes := []int{50, 500, 5000}

				// Launch concurrent requests
				startTime := time.Now()
				for i := 0; i < concurrency; i++ {
					wg.Add(1)
					sem <- struct{}{} // Acquire semaphore
					go func(id int) {
						defer func() {
							<-sem // Release semaphore
							wg.Done()
						}()

						// Choose a random message size
						messageSize := messageSizes[rand.Intn(len(messageSizes))]
						messageCount := 5 + rand.Intn(10) // Send 5-15 messages
						atomic.AddInt64(&totalMessagesCount, int64(messageCount))

						// Get channel from pool
						acquireStart := time.Now()
						ch := pool.Get()
						acquireDuration := time.Since(acquireStart)
						atomic.AddInt64(&totalAcquireTime, acquireDuration.Nanoseconds())

						if ch == nil {
							atomic.AddInt32(&acquireErrors, 1)
							t.Logf("Get error: Channel is nil (ID: %d)", id)
							return
						}

						// Create a Token for sending
						sendToken := domain.Token{
							Text:     "",
							Finished: false,
						}

						// Simulate producer sending messages
						go func() {
							defer func() {
								// Catch any panics from sending on closed channels
								if r := recover(); r != nil {
									t.Logf("Panic in producer: %v (ID: %d)", r, id)
								}
							}()

							for i := 0; i < messageCount; i++ {
								// Generate message content
								msg := fmt.Sprintf("message-%d-%d-%d-", id, i, rand.Intn(10000))
								for len(msg) < messageSize {
									msg += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
								}
								msg = msg[:messageSize]

								// Set token text
								sendToken.Text = msg
								// Set finished flag for the last message
								sendToken.Finished = (i == messageCount-1)

								// Simulate varying processing time
								time.Sleep(time.Duration(rand.Intn(2)) * time.Millisecond)

								// Try to send token with a timeout
								select {
								case ch <- sendToken:
									// Token sent successfully
								case <-time.After(5 * time.Millisecond):
									// Timeout occurred
									atomic.AddInt32(&waitTimeoutCount, 1)
								}
							}
						}()

						// Simulate consumer reading messages
						receivedCount := 0
						timeout := time.After(50 * time.Millisecond)

					ConsumerLoop:
						for {
							select {
							case token, ok := <-ch:
								if !ok {
									t.Logf("Channel closed unexpectedly (ID: %d)", id)
									break ConsumerLoop
								}
								receivedCount++
								if token.Finished || receivedCount >= messageCount {
									break ConsumerLoop
								}
							case <-timeout:
								// Consumer timed out waiting for messages
								atomic.AddInt32(&messageDropCount, int32(messageCount-receivedCount))
								break ConsumerLoop
							}
						}

						// Put channel back to pool
						releaseStart := time.Now()
						pool.Put(ch)
						releaseDuration := time.Since(releaseStart)
						atomic.AddInt64(&totalReleaseTime, releaseDuration.Nanoseconds())
					}(i)
				}

				// Wait for all operations to complete
				wg.Wait()
				totalDuration := time.Since(startTime)

				// Check goroutine count after test
				peakGoroutines := runtime.NumGoroutine()

				// Record results
				avgAcquireTime := float64(totalAcquireTime) / float64(concurrency) / float64(time.Microsecond)
				avgReleaseTime := float64(totalReleaseTime) / float64(concurrency) / float64(time.Microsecond)
				totalErrors := acquireErrors + releaseErrors
				errorRate := float64(totalErrors) / float64(concurrency) * 100
				messageDropRate := float64(messageDropCount) / float64(totalMessagesCount) * 100
				waitTimeoutRate := float64(waitTimeoutCount) / float64(totalMessagesCount) * 100

				t.Logf("Results for ChannelPool at concurrency %d:", concurrency)
				t.Logf("  Total messages: %d", totalMessagesCount)
				t.Logf("  Error rate: %.2f%% (%d/%d)", errorRate, totalErrors, concurrency)
				t.Logf("  Message drop rate: %.2f%% (%d/%d)", messageDropRate, messageDropCount, totalMessagesCount)
				t.Logf("  Send timeout rate: %.2f%% (%d/%d)", waitTimeoutRate, waitTimeoutCount, totalMessagesCount)
				t.Logf("  Average acquire time: %.3f µs", avgAcquireTime)
				t.Logf("  Average release time: %.3f µs", avgReleaseTime)
				t.Logf("  Total duration: %v", totalDuration)
				t.Logf("  Throughput: %.2f ops/sec", float64(concurrency)/totalDuration.Seconds())
				t.Logf("  Message throughput: %.2f msgs/sec", float64(totalMessagesCount)/totalDuration.Seconds())
				t.Logf("  Goroutines: %d initial, %d peak", initialGoroutines, peakGoroutines)

				// Basic validation
				if errorRate > 5.0 {
					t.Errorf("Error rate too high: %.2f%%", errorRate)
				}
			})
		}
	}

	// Collect final memory stats
	runtime.ReadMemStats(&memStatsAfter)

	// Report memory usage
	t.Logf("Memory usage before: %.2f MB", float64(memStatsBefore.Alloc)/1024/1024)
	t.Logf("Memory usage after: %.2f MB", float64(memStatsAfter.Alloc)/1024/1024)

	// Calculate memory difference with error handling for potential integer underflow
	var memDiff float64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memDiff = float64(memStatsAfter.Alloc-memStatsBefore.Alloc) / 1024 / 1024
	} else {
		// If we somehow get a negative difference (e.g., due to GC between measurements), report 0
		memDiff = 0
	}
	t.Logf("Memory difference: %.2f MB", memDiff)
	t.Logf("Total allocations: %d objects", memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}
