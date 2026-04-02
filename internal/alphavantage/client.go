// Package alphavantage provides a client for the Alpha Vantage API.
// It includes rate limiting, retry logic, and proper error handling.
package alphavantage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultRateLimit is the default requests per minute.
	DefaultRateLimit = 70
	// DefaultTimeout is the default request timeout.
	DefaultTimeout = 10 * time.Second
	// DefaultMaxRetries is the default maximum number of retry attempts.
	DefaultMaxRetries = 3
	// DefaultBaseDelay is the base delay for retry backoff.
	DefaultBaseDelay = time.Second
	// DefaultJitter is the maximum jitter for retry backoff.
	DefaultJitter = 200 * time.Millisecond
)

// APIError represents an error returned by the Alpha Vantage API.
type APIError struct {
	StatusCode int
	Message    string
	Endpoint   string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status=%d message=%q endpoint=%s", e.StatusCode, e.Message, e.Endpoint)
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
func (e *APIError) Unwrap() error {
	if e.StatusCode >= 400 && e.StatusCode < 500 {
		return errors.New("client error")
	}
	if e.StatusCode >= 500 {
		return errors.New("server error")
	}
	return nil
}

// Config holds the configuration for the Alpha Vantage client.
type Config struct {
	Key        string
	BaseURL    string
	RateLimit  int // Requests per minute
	Timeout    time.Duration
	MaxRetries int
	HTTPClient *http.Client // Optional, for testing
}

// Client is an Alpha Vantage API client with rate limiting and retry logic.
type Client struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	rateLimiter *tokenBucket
	config      Config
}

// New creates a new Alpha Vantage API client.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://www.alphavantage.co/query"
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = DefaultRateLimit
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = DefaultMaxRetries
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	return &Client{
		baseURL:     cfg.BaseURL,
		apiKey:      cfg.Key,
		httpClient:  cfg.HTTPClient,
		rateLimiter: newTokenBucket(cfg.RateLimit, time.Minute),
		config:      cfg,
	}
}

// get executes a GET request to the Alpha Vantage API with the given parameters.
// It handles rate limiting, retries on 5xx errors and timeouts, and returns the response body.
// The API key is automatically added to all requests and is never logged.
//
//nolint:gocyclo // Function complexity is reasonable for this use case
func (c *Client) get(ctx context.Context, params map[string]string) ([]byte, error) {
	// Build URL with parameters
	u, err := c.buildURL(params)
	if err != nil {
		return nil, fmt.Errorf("building request URL: %w", err)
	}

	var lastErr error
	var responseBody []byte

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Apply exponential backoff with jitter
			backoff := c.calculateBackoff(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, fmt.Errorf("context canceled during backoff: %w", ctx.Err())
			}
		}

		// Acquire rate limit token (blocks if budget exhausted)
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		// Create request with context
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
			// Retry on network errors and timeouts
			if c.shouldRetryError(err) && attempt < c.config.MaxRetries {
				continue
			}
			return nil, lastErr
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			// Retry on read errors
			if attempt < c.config.MaxRetries {
				continue
			}
			return nil, lastErr
		}

		// Check status code
		//nolint:gocritic // if-else chain is clearer for this error handling logic
		if resp.StatusCode == http.StatusTooManyRequests {
			// 429 - rate limit, retry with backoff
			lastErr = c.newAPIError(resp.StatusCode, "rate limited", u.Path)
			if attempt < c.config.MaxRetries {
				continue
			}
			// Exhausted retries, return error
			return nil, lastErr
		} else if resp.StatusCode >= 500 {
			// 5xx - server error, retry with backoff
			lastErr = c.newAPIError(resp.StatusCode, string(body), u.Path)
			if attempt < c.config.MaxRetries {
				continue
			}
			// Exhausted retries, return error
			return nil, lastErr
		} else if resp.StatusCode >= 400 {
			// 4xx (except 429) - client error, don't retry
			return nil, c.newAPIError(resp.StatusCode, c.extractErrorMessage(body), u.Path)
		}

		// Success - check for API-level errors in response body
		if apiErr := c.checkAPIError(body); apiErr != nil {
			return nil, apiErr
		}

		// Clear any previous error since this attempt succeeded
		lastErr = nil
		responseBody = body
		break
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return responseBody, nil
}

// buildURL constructs the full URL with query parameters.
// The API key is automatically added and is never logged.
func (c *Client) buildURL(params map[string]string) (*url.URL, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	// Ensure HTTPS for non-localhost URLs (for production security)
	// Allow HTTP for localhost/127.0.0.1 for testing
	if u.Scheme != "https" && u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" {
		u.Scheme = "https"
	}

	query := u.Query()
	query.Set("apikey", c.apiKey)

	for key, value := range params {
		query.Set(key, value)
	}

	u.RawQuery = query.Encode()
	return u, nil
}

// calculateBackoff calculates the exponential backoff delay with jitter.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: base * 2^(attempt-1)
	backoff := DefaultBaseDelay * time.Duration(1<<uint(attempt-1))

	// Add random jitter: ±200ms
	jitter := time.Duration(rand.Int63n(int64(DefaultJitter*2))) - DefaultJitter

	return backoff + jitter
}

// shouldRetryError determines if an error should trigger a retry.
func (c *Client) shouldRetryError(err error) bool {
	if err == nil {
		return false
	}

	// Retry on timeout errors
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Retry on network errors
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}

// newAPIError creates a new APIError with the given parameters.
func (c *Client) newAPIError(statusCode int, message, endpoint string) *APIError {
	// Truncate message if too long
	if len(message) > 500 {
		message = message[:500] + "..."
	}

	// Redact API key from message
	message = strings.ReplaceAll(message, c.apiKey, "***REDACTED***")

	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Endpoint:   endpoint,
	}
}

// extractErrorMessage attempts to extract a meaningful error message from the response body.
func (c *Client) extractErrorMessage(body []byte) string {
	// Try to parse as JSON
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err == nil {
		// Check for Alpha Vantage error message
		if errMsg, ok := jsonResp["Error Message"].(string); ok {
			return errMsg
		}
		// Check for API note
		if note, ok := jsonResp["Note"].(string); ok {
			return note
		}
		// Check for information field
		if info, ok := jsonResp["Information"].(string); ok {
			return info
		}
	}

	// Fallback to raw body (truncated)
	msg := string(body)
	if len(msg) > 500 {
		msg = msg[:500] + "..."
	}
	return msg
}

// checkAPIError checks if the response body contains an API-level error.
func (c *Client) checkAPIError(body []byte) error {
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return nil // Not valid JSON, assume success
	}

	// Alpha Vantage returns errors with specific keys
	if _, ok := jsonResp["Error Message"]; ok {
		return fmt.Errorf("API error: %s", jsonResp["Error Message"])
	}
	if note, ok := jsonResp["Note"].(string); ok {
		return fmt.Errorf("API note: %s", note)
	}
	if info, ok := jsonResp["Information"].(string); ok {
		return fmt.Errorf("API information: %s", info)
	}

	return nil
}

// tokenBucket implements a rate limiter using the token bucket algorithm.
type tokenBucket struct {
	mu           sync.Mutex
	tokens       float64
	capacity     float64
	refillRate   float64
	refillAmount float64
	lastRefill   time.Time
}

// newTokenBucket creates a new token bucket with the given capacity and refill interval.
func newTokenBucket(requestsPerMinute int, refillInterval time.Duration) *tokenBucket {
	return &tokenBucket{
		capacity:     float64(requestsPerMinute),
		tokens:       float64(requestsPerMinute), // Start with full bucket
		refillRate:   float64(requestsPerMinute) / refillInterval.Seconds(),
		refillAmount: float64(requestsPerMinute) / float64(refillInterval/time.Second),
		lastRefill:   time.Now(),
	}
}

// Wait blocks until a token is available or the context is canceled.
func (tb *tokenBucket) Wait(ctx context.Context) error {
	for {
		tb.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(tb.lastRefill)

		// Refill tokens based on elapsed time
		if elapsed > 0 {
			tb.tokens += tb.refillRate * elapsed.Seconds()
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			tb.lastRefill = now
		}

		if tb.tokens >= 1 {
			tb.tokens--
			tb.mu.Unlock()
			return nil
		}

		// Calculate wait time for next token
		waitTime := time.Duration((1 - tb.tokens) / tb.refillRate * float64(time.Second))
		tb.mu.Unlock()

		select {
		case <-time.After(waitTime):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// String returns a string representation of the token bucket state.
func (tb *tokenBucket) String() string {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return fmt.Sprintf("tokenBucket{tokens: %.2f/%.2f}", tb.tokens, tb.capacity)
}

// Key returns the API key (for testing purposes only).
// In production, this should never be called.
func (c *Client) Key() string {
	return c.apiKey
}

// BaseURL returns the base URL (for testing purposes only).
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Config returns the client configuration (for testing purposes only).
func (c *Client) Config() Config {
	return c.config
}
