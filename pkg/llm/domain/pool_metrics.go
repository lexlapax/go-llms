// ABOUTME: Metrics-enabled object pools for performance monitoring
// ABOUTME: Tracks pool utilization, allocation rates, and memory usage

package domain

import (
	"time"

	"github.com/lexlapax/go-llms/pkg/util/metrics"
)

// ResponsePoolWithMetrics extends ResponsePool with metrics.
// It tracks allocation patterns, timing, and pool utilization
// for performance monitoring and optimization.
type ResponsePoolWithMetrics struct {
	ResponsePool
	metrics *metrics.PoolMetrics
}

// NewResponsePoolWithMetrics creates a new response pool with metrics.
// The pool tracks allocation and return counts, timing data,
// and pool utilization statistics.
//
// Returns a new ResponsePoolWithMetrics instance.
func NewResponsePoolWithMetrics() *ResponsePoolWithMetrics {
	return &ResponsePoolWithMetrics{
		ResponsePool: *NewResponsePool(),
		metrics:      metrics.NewPoolMetrics("response_pool"),
	}
}

// Get retrieves a Response from the pool with metrics.
// Records allocation timing and increments allocation counter.
//
// Returns a Response object ready for use.
func (p *ResponsePoolWithMetrics) Get() *Response {
	startTime := time.Now()

	// Either create a new object or get one from the pool
	resp := p.ResponsePool.Get()

	// Record metrics
	p.metrics.RecordAllocation()
	p.metrics.RecordAllocationTime(time.Since(startTime))

	return resp
}

// Put returns a Response to the pool after use with metrics.
// Records the return operation for pool utilization tracking.
//
// Parameters:
//   - resp: The Response to return to the pool
func (p *ResponsePoolWithMetrics) Put(resp *Response) {
	if resp == nil {
		return
	}

	// Return to the pool
	p.ResponsePool.Put(resp)

	// Record metrics
	p.metrics.RecordReturn()
}

// NewResponse creates a new Response with the given content using the pool with metrics.
// Tracks the complete allocation lifecycle including timing.
//
// Parameters:
//   - content: The content for the response
//
// Returns a Response with the specified content.
func (p *ResponsePoolWithMetrics) NewResponse(content string) Response {
	startTime := time.Now()

	// Get from pool
	resp := p.Get()
	resp.Content = content

	// Create a copy to return by value (Response, not *Response)
	result := *resp

	// Return the object to the pool
	p.Put(resp)

	// Record allocation time
	p.metrics.RecordAllocationTime(time.Since(startTime))

	return result
}

// GetMetrics returns the pool size, allocation count, and return count.
// Provides snapshot of current pool utilization statistics.
//
// Returns:
//   - size: Current pool size
//   - allocated: Total objects allocated
//   - returned: Total objects returned
func (p *ResponsePoolWithMetrics) GetMetrics() (size int64, allocated int64, returned int64) {
	allocated, returned = p.metrics.GetAllocationCount()
	size = p.metrics.GetPoolSize()
	return
}

// GetAverageAllocationTime returns the average time to allocate a new object.
//
// Returns the average allocation duration.
func (p *ResponsePoolWithMetrics) GetAverageAllocationTime() time.Duration {
	return p.metrics.GetAverageAllocationTime()
}

// GetAverageWaitTime returns the average time waiting for an object.
//
// Returns the average wait duration.
func (p *ResponsePoolWithMetrics) GetAverageWaitTime() time.Duration {
	return p.metrics.GetAverageWaitTime()
}

// TokenPoolWithMetrics extends TokenPool with metrics.
// Provides detailed performance monitoring for token allocation
// patterns during streaming operations.
type TokenPoolWithMetrics struct {
	TokenPool
	metrics *metrics.PoolMetrics
}

// NewTokenPoolWithMetrics creates a new token pool with metrics.
// Enables monitoring of token allocation performance for optimization.
//
// Returns a new TokenPoolWithMetrics instance.
func NewTokenPoolWithMetrics() *TokenPoolWithMetrics {
	return &TokenPoolWithMetrics{
		TokenPool: *NewTokenPool(),
		metrics:   metrics.NewPoolMetrics("token_pool"),
	}
}

// Get retrieves a Token from the pool with metrics.
// Records allocation timing and increments allocation counter.
//
// Returns a Token object ready for use.
func (p *TokenPoolWithMetrics) Get() *Token {
	startTime := time.Now()

	// Either create a new object or get one from the pool
	token := p.TokenPool.Get()

	// Record metrics
	p.metrics.RecordAllocation()
	p.metrics.RecordAllocationTime(time.Since(startTime))

	return token
}

// Put returns a Token to the pool after use with metrics.
// Records the return operation for pool utilization tracking.
//
// Parameters:
//   - token: The Token to return to the pool
func (p *TokenPoolWithMetrics) Put(token *Token) {
	if token == nil {
		return
	}

	// Return to the pool
	p.TokenPool.Put(token)

	// Record metrics
	p.metrics.RecordReturn()
}

// NewToken creates a new Token with the given text and finished flag using the pool with metrics.
// Tracks the complete token creation lifecycle including timing.
//
// Parameters:
//   - text: The token text
//   - finished: Whether this token represents the end of stream
//
// Returns a Token with the specified properties.
func (p *TokenPoolWithMetrics) NewToken(text string, finished bool) Token {
	startTime := time.Now()

	// Get from pool
	token := p.Get()
	token.Text = text
	token.Finished = finished

	// Create a copy to return by value (Token, not *Token)
	result := *token

	// Return the object to the pool
	p.Put(token)

	// Record allocation time
	p.metrics.RecordAllocationTime(time.Since(startTime))

	return result
}

// GetMetrics returns the pool size, allocation count, and return count.
// Provides snapshot of current pool utilization statistics.
//
// Returns:
//   - size: Current pool size
//   - allocated: Total tokens allocated
//   - returned: Total tokens returned
func (p *TokenPoolWithMetrics) GetMetrics() (size int64, allocated int64, returned int64) {
	allocated, returned = p.metrics.GetAllocationCount()
	size = p.metrics.GetPoolSize()
	return
}

// GetAverageAllocationTime returns the average time to allocate a new object.
//
// Returns the average allocation duration.
func (p *TokenPoolWithMetrics) GetAverageAllocationTime() time.Duration {
	return p.metrics.GetAverageAllocationTime()
}

// GetAverageWaitTime returns the average time waiting for an object.
//
// Returns the average wait duration.
func (p *TokenPoolWithMetrics) GetAverageWaitTime() time.Duration {
	return p.metrics.GetAverageWaitTime()
}

// ChannelPoolWithMetrics extends ChannelPool with metrics.
// Monitors channel allocation patterns for streaming performance
// analysis and optimization.
type ChannelPoolWithMetrics struct {
	ChannelPool
	metrics *metrics.PoolMetrics
}

// NewChannelPoolWithMetrics creates a new channel pool with metrics.
// Enables monitoring of channel allocation for streaming operations.
//
// Returns a new ChannelPoolWithMetrics instance.
func NewChannelPoolWithMetrics() *ChannelPoolWithMetrics {
	return &ChannelPoolWithMetrics{
		ChannelPool: *NewChannelPool(),
		metrics:     metrics.NewPoolMetrics("channel_pool"),
	}
}

// Get retrieves a channel from the pool with metrics.
// Records allocation timing and increments allocation counter.
//
// Returns a chan Token ready for use.
func (p *ChannelPoolWithMetrics) Get() chan Token {
	startTime := time.Now()

	// Either create a new channel or get one from the pool
	ch := p.ChannelPool.Get()

	// Record metrics
	p.metrics.RecordAllocation()
	p.metrics.RecordAllocationTime(time.Since(startTime))

	return ch
}

// Put returns a channel to the pool after use with metrics.
// Records the return operation for pool utilization tracking.
//
// Parameters:
//   - ch: The channel to return to the pool
func (p *ChannelPoolWithMetrics) Put(ch chan Token) {
	if ch == nil {
		return
	}

	// Return to the pool
	p.ChannelPool.Put(ch)

	// Record metrics
	p.metrics.RecordReturn()
}

// GetResponseStream creates a new response stream using the pool with metrics.
// Returns both read-only and write interfaces to the same channel.
//
// Returns:
//   - ResponseStream: Read-only channel for consuming tokens
//   - chan Token: Write channel for producing tokens
func (p *ChannelPoolWithMetrics) GetResponseStream() (ResponseStream, chan Token) {
	ch := p.Get()
	return ch, ch
}

// GetMetrics returns the pool size, allocation count, and return count.
// Provides snapshot of current pool utilization statistics.
//
// Returns:
//   - size: Current pool size
//   - allocated: Total channels allocated
//   - returned: Total channels returned
func (p *ChannelPoolWithMetrics) GetMetrics() (size int64, allocated int64, returned int64) {
	allocated, returned = p.metrics.GetAllocationCount()
	size = p.metrics.GetPoolSize()
	return
}

// GetAverageAllocationTime returns the average time to allocate a new channel.
//
// Returns the average allocation duration.
func (p *ChannelPoolWithMetrics) GetAverageAllocationTime() time.Duration {
	return p.metrics.GetAverageAllocationTime()
}

// GetAverageWaitTime returns the average time waiting for a channel.
//
// Returns the average wait duration.
func (p *ChannelPoolWithMetrics) GetAverageWaitTime() time.Duration {
	return p.metrics.GetAverageWaitTime()
}
