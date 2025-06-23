package domain

// ABOUTME: Object pool implementations for Response and Message types
// ABOUTME: Reduces memory allocations through efficient object reuse

import (
	"sync"
	"unsafe"
)

// ResponsePool is a pool of Response objects that can be reused to reduce memory allocations.
// It implements sync.Pool to efficiently manage Response object lifecycle,
// reducing GC pressure in high-throughput scenarios.
type ResponsePool struct {
	pool sync.Pool
}

// Global singleton response pool
var (
	globalResponsePool *ResponsePool
	responsePoolOnce   sync.Once
)

// GetResponsePool returns the singleton global response pool.
// Uses sync.Once to ensure thread-safe singleton initialization.
//
// Returns the global ResponsePool instance.
func GetResponsePool() *ResponsePool {
	responsePoolOnce.Do(func() {
		globalResponsePool = NewResponsePool()
	})
	return globalResponsePool
}

// NewResponsePool creates a new response pool.
// The pool automatically creates new Response objects when empty.
//
// Returns a new ResponsePool instance.
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

// Get retrieves a Response from the pool.
// Returns a reused Response object or creates a new one if pool is empty.
//
// Returns a Response object ready for use.
func (p *ResponsePool) Get() *Response {
	return p.pool.Get().(*Response)
}

// Put returns a Response to the pool after use.
// Automatically clears content to prevent data leaks between uses.
// Uses optimized clearing for large content to minimize allocations.
//
// Parameters:
//   - resp: The Response to return to the pool (nil is ignored)
func (p *ResponsePool) Put(resp *Response) {
	if resp == nil {
		return
	}

	// Clear the Response fields before returning to the pool
	// For large content, this is optimized to minimize allocations
	if len(resp.Content) > 1024 {
		// Use optimized clearing for large content
		// This just changes the length to 0 without allocation
		ZeroString(&resp.Content)
	} else {
		// For smaller content, simple assignment is faster
		resp.Content = ""
	}

	p.pool.Put(resp)
}

// NewResponse creates a new Response with the given content using the pool.
// This is a convenience method that gets a Response from the pool,
// sets its content, and returns it by value.
//
// Parameters:
//   - content: The content for the response
//
// Returns a Response with the specified content.
func (p *ResponsePool) NewResponse(content string) Response {
	resp := p.Get()
	resp.Content = content

	// Create a copy to return by value (Response, not *Response)
	result := *resp

	// Return the object to the pool
	p.Put(resp)

	return result
}

// TokenPool is a pool of Token objects that can be reused to reduce memory allocations.
// Manages Token object lifecycle for streaming operations,
// reducing allocation overhead in high-frequency token generation.
type TokenPool struct {
	pool sync.Pool
}

// Global singleton token pool
var (
	globalTokenPool *TokenPool
	tokenPoolOnce   sync.Once
)

// GetTokenPool returns the singleton global token pool.
// Uses sync.Once to ensure thread-safe singleton initialization.
//
// Returns the global TokenPool instance.
func GetTokenPool() *TokenPool {
	tokenPoolOnce.Do(func() {
		globalTokenPool = NewTokenPool()
	})
	return globalTokenPool
}

// NewTokenPool creates a new token pool.
// The pool automatically creates new Token objects when empty.
//
// Returns a new TokenPool instance.
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

// Get retrieves a Token from the pool.
// Returns a reused Token object or creates a new one if pool is empty.
//
// Returns a Token object ready for use.
func (p *TokenPool) Get() *Token {
	return p.pool.Get().(*Token)
}

// Put returns a Token to the pool after use.
// Automatically clears token data to prevent information leaks.
// Uses optimized clearing for large text content.
//
// Parameters:
//   - token: The Token to return to the pool (nil is ignored)
func (p *TokenPool) Put(token *Token) {
	if token == nil {
		return
	}

	// Clear the Token fields before returning to the pool
	// For large content, this is optimized to minimize allocations
	if len(token.Text) > 1024 {
		// Use optimized clearing for large content
		ZeroString(&token.Text)
	} else {
		// For smaller content, simple assignment is faster
		token.Text = ""
	}

	token.Finished = false
	p.pool.Put(token)
}

// NewToken creates a new Token with the given text and finished flag using the pool.
// This is a convenience method that gets a Token from the pool,
// sets its properties, and returns it by value.
//
// Parameters:
//   - text: The token text
//   - finished: Whether this token represents the end of stream
//
// Returns a Token with the specified properties.
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

// ChannelPoolSize is the default buffer size for channels from the pool
const ChannelPoolSize = 20

// ChannelPool is a pool of channels that can be reused to reduce memory allocations.
// This significantly reduces GC pressure in streaming operations by reusing
// buffered channels instead of creating new ones for each stream.
type ChannelPool struct {
	pool sync.Pool
}

// Global singleton channel pool
var (
	globalChannelPool *ChannelPool
	channelPoolOnce   sync.Once
)

// GetChannelPool returns the singleton global channel pool.
// Uses sync.Once to ensure thread-safe singleton initialization.
//
// Returns the global ChannelPool instance.
func GetChannelPool() *ChannelPool {
	channelPoolOnce.Do(func() {
		globalChannelPool = NewChannelPool()
	})
	return globalChannelPool
}

// NewChannelPool creates a new channel pool.
// Creates buffered channels with ChannelPoolSize capacity to prevent blocking.
//
// Returns a new ChannelPool instance.
func NewChannelPool() *ChannelPool {
	return &ChannelPool{
		pool: sync.Pool{
			New: func() interface{} {
				// Create a new channel when the pool is empty
				// Use a buffered channel with a reasonable size to prevent blocking
				return make(chan Token, ChannelPoolSize)
			},
		},
	}
}

// Get retrieves a channel from the pool.
// Returns a reused buffered channel or creates a new one if pool is empty.
//
// Returns a chan Token ready for use.
func (p *ChannelPool) Get() chan Token {
	return p.pool.Get().(chan Token)
}

// Put returns a channel to the pool after use.
// Automatically drains any remaining tokens to ensure the channel is clean.
// Closed channels are not returned to the pool.
//
// Parameters:
//   - ch: The channel to return to the pool (nil is ignored)
func (p *ChannelPool) Put(ch chan Token) {
	if ch == nil {
		return
	}

	// Drain any remaining tokens to ensure the channel is empty
	// This is a non-blocking operation
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				// Channel is closed, don't put it back
				return
			}
		default:
			// Channel is empty
			p.pool.Put(ch)
			return
		}
	}
}

// GetResponseStream creates a new response stream using the pool.
// Returns both a read-only ResponseStream and the underlying channel
// for writing. The caller must close the channel when done.
//
// Returns:
//   - ResponseStream: Read-only channel for consuming tokens
//   - chan Token: Write channel for producing tokens
func (p *ChannelPool) GetResponseStream() (ResponseStream, chan Token) {
	ch := p.Get()
	return ch, ch
}

// StringHeader represents the internal header structure of a string.
// This is equivalent to the reflect.StringHeader structure and is used
// for unsafe string operations to avoid allocations.
type StringHeader struct {
	Data uintptr
	Len  int
}

// ZeroString clears a string's content without allocation.
// This is an unsafe operation that manipulates the string header
// to point to an empty string, avoiding memory allocation.
// Use with caution and only when performance is critical.
//
// Parameters:
//   - s: Pointer to the string to clear
func ZeroString(s *string) {
	if s == nil || *s == "" {
		return
	}

	// Create a new empty string
	empty := ""

	// Get the string headers
	emptyHeader := (*StringHeader)(unsafe.Pointer(&empty))
	sHeader := (*StringHeader)(unsafe.Pointer(s))

	// Make the target string point to the empty string data
	sHeader.Data = emptyHeader.Data
	sHeader.Len = 0
}
