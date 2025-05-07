package domain

import (
	"sync"
)

// ResponsePool is a pool of Response objects that can be reused to reduce memory allocations
type ResponsePool struct {
	pool sync.Pool
}

// Global singleton response pool
var (
	globalResponsePool *ResponsePool
	responsePoolOnce   sync.Once
)

// GetResponsePool returns the singleton global response pool
func GetResponsePool() *ResponsePool {
	responsePoolOnce.Do(func() {
		globalResponsePool = NewResponsePool()
	})
	return globalResponsePool
}

// NewResponsePool creates a new response pool
func NewResponsePool() *ResponsePool {
	return &ResponsePool{
		pool: sync.Pool{
			New: func() interface{} {
				// Create a new Response when the pool is empty
				return &Response{}
			},
		},
	}
}

// Get retrieves a Response from the pool
func (p *ResponsePool) Get() *Response {
	return p.pool.Get().(*Response)
}

// Put returns a Response to the pool after use
// Make sure to clear any sensitive data before putting a Response back
func (p *ResponsePool) Put(resp *Response) {
	if resp == nil {
		return
	}
	
	// Clear the Response fields before returning to the pool
	resp.Content = ""
	
	p.pool.Put(resp)
}

// NewResponse creates a new Response with the given content using the pool
func (p *ResponsePool) NewResponse(content string) Response {
	resp := p.Get()
	resp.Content = content
	
	// Create a copy to return by value (Response, not *Response)
	result := *resp
	
	// Return the object to the pool
	p.Put(resp)
	
	return result
}

// TokenPool is a pool of Token objects that can be reused to reduce memory allocations
type TokenPool struct {
	pool sync.Pool
}

// Global singleton token pool
var (
	globalTokenPool *TokenPool
	tokenPoolOnce   sync.Once
)

// GetTokenPool returns the singleton global token pool
func GetTokenPool() *TokenPool {
	tokenPoolOnce.Do(func() {
		globalTokenPool = NewTokenPool()
	})
	return globalTokenPool
}

// NewTokenPool creates a new token pool
func NewTokenPool() *TokenPool {
	return &TokenPool{
		pool: sync.Pool{
			New: func() interface{} {
				// Create a new Token when the pool is empty
				return &Token{}
			},
		},
	}
}

// Get retrieves a Token from the pool
func (p *TokenPool) Get() *Token {
	return p.pool.Get().(*Token)
}

// Put returns a Token to the pool after use
func (p *TokenPool) Put(token *Token) {
	if token == nil {
		return
	}
	
	// Clear the Token fields before returning to the pool
	token.Text = ""
	token.Finished = false
	
	p.pool.Put(token)
}

// NewToken creates a new Token with the given text and finished flag using the pool
func (p *TokenPool) NewToken(text string, finished bool) Token {
	token := p.Get()
	token.Text = text
	token.Finished = finished
	
	// Create a copy to return by value (Token, not *Token)
	result := *token
	
	// Return the object to the pool
	p.Put(token)
	
	return result
}