package domain

import (
	"time"

	"github.com/lexlapax/go-llms/pkg/util/metrics"
)

// ResponsePoolWithMetrics extends ResponsePool with metrics
type ResponsePoolWithMetrics struct {
	ResponsePool
	metrics *metrics.PoolMetrics
}

// NewResponsePoolWithMetrics creates a new response pool with metrics
func NewResponsePoolWithMetrics() *ResponsePoolWithMetrics {
	return &ResponsePoolWithMetrics{
		ResponsePool: *NewResponsePool(),
		metrics:      metrics.NewPoolMetrics("response_pool"),
	}
}

// Get retrieves a Response from the pool with metrics
func (p *ResponsePoolWithMetrics) Get() *Response {
	startTime := time.Now()
	
	// Either create a new object or get one from the pool
	resp := p.ResponsePool.Get()
	
	// Record metrics
	p.metrics.RecordAllocation()
	p.metrics.RecordAllocationTime(time.Since(startTime))
	
	return resp
}

// Put returns a Response to the pool after use with metrics
func (p *ResponsePoolWithMetrics) Put(resp *Response) {
	if resp == nil {
		return
	}
	
	// Return to the pool
	p.ResponsePool.Put(resp)
	
	// Record metrics
	p.metrics.RecordReturn()
}

// NewResponse creates a new Response with the given content using the pool with metrics
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

// GetMetrics returns the pool size, allocation count, and return count
func (p *ResponsePoolWithMetrics) GetMetrics() (size int64, allocated int64, returned int64) {
	allocated, returned = p.metrics.GetAllocationCount()
	size = p.metrics.GetPoolSize()
	return
}

// GetAverageAllocationTime returns the average time to allocate a new object
func (p *ResponsePoolWithMetrics) GetAverageAllocationTime() time.Duration {
	return p.metrics.GetAverageAllocationTime()
}

// GetAverageWaitTime returns the average time waiting for an object
func (p *ResponsePoolWithMetrics) GetAverageWaitTime() time.Duration {
	return p.metrics.GetAverageWaitTime()
}

// TokenPoolWithMetrics extends TokenPool with metrics
type TokenPoolWithMetrics struct {
	TokenPool
	metrics *metrics.PoolMetrics
}

// NewTokenPoolWithMetrics creates a new token pool with metrics
func NewTokenPoolWithMetrics() *TokenPoolWithMetrics {
	return &TokenPoolWithMetrics{
		TokenPool: *NewTokenPool(),
		metrics:   metrics.NewPoolMetrics("token_pool"),
	}
}

// Get retrieves a Token from the pool with metrics
func (p *TokenPoolWithMetrics) Get() *Token {
	startTime := time.Now()
	
	// Either create a new object or get one from the pool
	token := p.TokenPool.Get()
	
	// Record metrics
	p.metrics.RecordAllocation()
	p.metrics.RecordAllocationTime(time.Since(startTime))
	
	return token
}

// Put returns a Token to the pool after use with metrics
func (p *TokenPoolWithMetrics) Put(token *Token) {
	if token == nil {
		return
	}
	
	// Return to the pool
	p.TokenPool.Put(token)
	
	// Record metrics
	p.metrics.RecordReturn()
}

// NewToken creates a new Token with the given text and finished flag using the pool with metrics
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

// GetMetrics returns the pool size, allocation count, and return count
func (p *TokenPoolWithMetrics) GetMetrics() (size int64, allocated int64, returned int64) {
	allocated, returned = p.metrics.GetAllocationCount()
	size = p.metrics.GetPoolSize()
	return
}

// GetAverageAllocationTime returns the average time to allocate a new object
func (p *TokenPoolWithMetrics) GetAverageAllocationTime() time.Duration {
	return p.metrics.GetAverageAllocationTime()
}

// GetAverageWaitTime returns the average time waiting for an object
func (p *TokenPoolWithMetrics) GetAverageWaitTime() time.Duration {
	return p.metrics.GetAverageWaitTime()
}

// ChannelPoolWithMetrics extends ChannelPool with metrics
type ChannelPoolWithMetrics struct {
	ChannelPool
	metrics *metrics.PoolMetrics
}

// NewChannelPoolWithMetrics creates a new channel pool with metrics
func NewChannelPoolWithMetrics() *ChannelPoolWithMetrics {
	return &ChannelPoolWithMetrics{
		ChannelPool: *NewChannelPool(),
		metrics:     metrics.NewPoolMetrics("channel_pool"),
	}
}

// Get retrieves a channel from the pool with metrics
func (p *ChannelPoolWithMetrics) Get() chan Token {
	startTime := time.Now()
	
	// Either create a new channel or get one from the pool
	ch := p.ChannelPool.Get()
	
	// Record metrics
	p.metrics.RecordAllocation()
	p.metrics.RecordAllocationTime(time.Since(startTime))
	
	return ch
}

// Put returns a channel to the pool after use with metrics
func (p *ChannelPoolWithMetrics) Put(ch chan Token) {
	if ch == nil {
		return
	}
	
	// Return to the pool
	p.ChannelPool.Put(ch)
	
	// Record metrics
	p.metrics.RecordReturn()
}

// GetResponseStream creates a new response stream using the pool with metrics
func (p *ChannelPoolWithMetrics) GetResponseStream() (ResponseStream, chan Token) {
	ch := p.Get()
	return ch, ch
}

// GetMetrics returns the pool size, allocation count, and return count
func (p *ChannelPoolWithMetrics) GetMetrics() (size int64, allocated int64, returned int64) {
	allocated, returned = p.metrics.GetAllocationCount()
	size = p.metrics.GetPoolSize()
	return
}

// GetAverageAllocationTime returns the average time to allocate a new channel
func (p *ChannelPoolWithMetrics) GetAverageAllocationTime() time.Duration {
	return p.metrics.GetAverageAllocationTime()
}

// GetAverageWaitTime returns the average time waiting for a channel
func (p *ChannelPoolWithMetrics) GetAverageWaitTime() time.Duration {
	return p.metrics.GetAverageWaitTime()
}