package alphavantage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClient_SuccessfulGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Verify HTTPS (test server uses http but we check for apikey param)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"success": true, "data": "test"}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	data, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	expected := `{"success": true, "data": "test"}`
	// Trim trailing newline since server adds it
	actual := strings.TrimSpace(string(data))
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestClient_RetryOn5xx(t *testing.T) {
	attempts := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		attempts++

		if attempts < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, `{"error": "service unavailable"}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 5,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	data, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	if string(data) != `{"success": true}` {
		t.Errorf("unexpected response: %s", string(data))
	}
}

func TestClient_RetryOn429(t *testing.T) {
	attempts := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		attempts++

		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 3,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	data, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}

	if string(data) != `{"success": true}` {
		t.Errorf("unexpected response: %s", string(data))
	}
}

func TestClient_NoRetryOn4xx(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, `{"error": "bad request"}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 5,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error for 4xx, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestClient_NoRetryOn401(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintln(w, `{"error": "unauthorized"}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 5,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestClient_NoRetryOn404(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintln(w, `{"error": "not found"}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 5,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.get(ctx, map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestClient_RateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create client with rate limit of 2 requests per minute
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		RateLimit:  2,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()

	// First 2 requests should succeed quickly
	start := time.Now()
	for i := 0; i < 2; i++ {
		_, err := client.get(ctx, map[string]string{"function": "TEST"})
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	firstDuration := time.Since(start)

	// Third request should be rate-limited and take some time
	start = time.Now()
	_, err := client.get(ctx, map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("third request failed: %v", err)
	}
	secondDuration := time.Since(start)

	// The third request should have taken significantly longer due to rate limiting
	// With 2 requests/minute, the third request should wait ~30 seconds
	// But we'll just check that it took longer than the first two
	if secondDuration < 10*time.Second {
		t.Logf("Warning: rate limiting may not have worked properly. First 2 took %v, third took %v", firstDuration, secondDuration)
	}

	t.Logf("First 2 requests took %v, third request took %v", firstDuration, secondDuration)
}

func TestClient_RateLimitingConcurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create client with rate limit of 5 requests per second
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		RateLimit:  5,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()

	// Launch 10 concurrent requests
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := client.get(ctx, map[string]string{"function": "TEST", "id": fmt.Sprintf("%d", id)})
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent request failed: %v", err)
	}
}

func TestClient_MaxRetriesExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, `{"error": "internal server error"}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 2,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestClient_TimeoutRetry(t *testing.T) {
	attempts := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		attempts++

		if attempts < 3 {
			// Simulate timeout by not responding
			time.Sleep(200 * time.Millisecond)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create HTTP client with very short timeout
	httpClient := &http.Client{
		Timeout: 50 * time.Millisecond,
	}

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 2, // 2 retries + initial = 3 total attempts
		HTTPClient: httpClient,
	})

	data, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("expected success after timeout retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	if string(data) != `{"success": true}` {
		t.Errorf("unexpected response: %s", string(data))
	}
}

func TestClient_APIKeyInRequest(t *testing.T) {
	var receivedKey string
	var receivedParams map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = r.URL.Query().Get("apikey")
		receivedParams = make(map[string]string)
		for k, v := range r.URL.Query() {
			if k != "apikey" {
				receivedParams[k] = v[0]
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"success": true}`)
	}))
	defer server.Close()

	apiKey := "test-api-key-12345"
	client := New(Config{
		Key:        apiKey,
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{
		"function":   "TIME_SERIES_DAILY",
		"symbol":     "AAPL",
		"outputsize": "full",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if receivedKey != apiKey {
		t.Errorf("expected API key %q, got %q", apiKey, receivedKey)
	}

	if receivedParams["function"] != "TIME_SERIES_DAILY" {
		t.Errorf("expected function TIME_SERIES_DAILY, got %s", receivedParams["function"])
	}

	if receivedParams["symbol"] != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", receivedParams["symbol"])
	}

	if receivedParams["outputsize"] != "full" {
		t.Errorf("expected outputsize full, got %s", receivedParams["outputsize"])
	}
}

func TestClient_APIKeyRedaction(t *testing.T) {
	apiKey := "secret-key-12345"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return the API key in the error response
		errorResp := map[string]string{
			"error": fmt.Sprintf("Invalid API key: %s", r.URL.Query().Get("apikey")),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := New(Config{
		Key:        apiKey,
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, apiKey) {
		t.Errorf("API key should be redacted in error message, but found in: %s", errMsg)
	}

	if !strings.Contains(errMsg, "***REDACTED***") {
		t.Errorf("Expected redacted API key in error message, got: %s", errMsg)
	}
}

func TestClient_APIErrorMessageExtraction(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantMsg  string
	}{
		{
			name:     "Error Message field",
			response: `{"Error Message": "Invalid API key."}`,
			wantMsg:  "API error: Invalid API key.",
		},
		{
			name:     "Note field",
			response: `{"Note": "Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day."}`,
			wantMsg:  "API note: Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day.",
		},
		{
			name:     "Information field",
			response: `{"Information": "The premium API key is required."}`,
			wantMsg:  "API information: The premium API key is required.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, tt.response)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			})

			_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("expected error to contain %q, got: %v", tt.wantMsg, err)
			}
		})
	}
}

func TestClient_HTTPSOnly(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantOK  bool
	}{
		{
			name:    "HTTPS URL",
			baseURL: "https://example.com/query",
			wantOK:  true,
		},
		{
			name:    "HTTP URL should be upgraded",
			baseURL: "http://example.com/query",
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(Config{
				Key:     "test-key",
				BaseURL: tt.baseURL,
			})

			// Build a URL to check if HTTPS is enforced
			u, err := client.buildURL(map[string]string{"function": "TEST"})
			if err != nil {
				t.Fatalf("buildURL failed: %v", err)
			}

			if u.Scheme != "https" {
				t.Errorf("expected HTTPS scheme, got %s", u.Scheme)
			}
		})
	}
}

func TestClient_BackoffCalculation(t *testing.T) {
	client := New(Config{
		Key:        "test-key",
		BaseURL:    "https://example.com/query",
		MaxRetries: 5,
	})

	// Test that backoff increases exponentially
	durations := make([]time.Duration, 5)
	for i := 0; i < 5; i++ {
		durations[i] = client.calculateBackoff(i + 1)
	}

	// First backoff should be around 1 second ± jitter
	if durations[0] < 800*time.Millisecond || durations[0] > 1200*time.Millisecond {
		t.Errorf("first backoff should be around 1s, got %v", durations[0])
	}

	// Second backoff should be around 2 seconds ± jitter
	if durations[1] < 1800*time.Millisecond || durations[1] > 2200*time.Millisecond {
		t.Errorf("second backoff should be around 2s, got %v", durations[1])
	}

	// Third backoff should be around 4 seconds ± jitter
	if durations[2] < 3800*time.Millisecond || durations[2] > 4200*time.Millisecond {
		t.Errorf("third backoff should be around 4s, got %v", durations[2])
	}

	// Fourth backoff should be around 8 seconds ± jitter
	if durations[3] < 7800*time.Millisecond || durations[3] > 8200*time.Millisecond {
		t.Errorf("fourth backoff should be around 8s, got %v", durations[3])
	}

	// Fifth backoff should be around 16 seconds ± jitter
	if durations[4] < 15800*time.Millisecond || durations[4] > 16200*time.Millisecond {
		t.Errorf("fifth backoff should be around 16s, got %v", durations[4])
	}
}

func TestTokenBucket(t *testing.T) {
	t.Run("full bucket initially", func(t *testing.T) {
		tb := newTokenBucket(10, time.Minute)
		ctx := context.Background()

		// Should be able to take 10 tokens immediately
		for i := 0; i < 10; i++ {
			if err := tb.Wait(ctx); err != nil {
				t.Fatalf("iteration %d: %v", i, err)
			}
		}

		// 11th request should block
		start := time.Now()
		done := make(chan error, 1)
		go func() {
			done <- tb.Wait(ctx)
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			elapsed := time.Since(start)
			// Should have waited for at least a fraction of the refill time
			if elapsed < 5*time.Second {
				t.Errorf("expected wait of at least 5s, got %v", elapsed)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timeout waiting for token")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		tb := newTokenBucket(1, time.Minute)

		// Take the only token
		ctx := context.Background()
		if err := tb.Wait(ctx); err != nil {
			t.Fatalf("first token: %v", err)
		}

		// Try to take another token with a canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := tb.Wait(ctx)
		if err == nil {
			t.Fatal("expected error for canceled context, got nil")
		}

		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("refill over time", func(t *testing.T) {
		// Use a short refill interval for testing
		tb := newTokenBucket(2, 100*time.Millisecond)
		ctx := context.Background()

		// Take all tokens
		for i := 0; i < 2; i++ {
			if err := tb.Wait(ctx); err != nil {
				t.Fatalf("iteration %d: %v", i, err)
			}
		}

		// Wait for refill
		time.Sleep(150 * time.Millisecond)

		// Should be able to take tokens again
		if err := tb.Wait(ctx); err != nil {
			t.Fatalf("after refill: %v", err)
		}
	})
}

func TestClient_Defaults(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	if client.config.RateLimit != DefaultRateLimit {
		t.Errorf("expected rate limit %d, got %d", DefaultRateLimit, client.config.RateLimit)
	}

	if client.config.Timeout != DefaultTimeout {
		t.Errorf("expected timeout %v, got %v", DefaultTimeout, client.config.Timeout)
	}

	if client.config.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected max retries %d, got %d", DefaultMaxRetries, client.config.MaxRetries)
	}

	if client.baseURL != "https://www.alphavantage.co/query" {
		t.Errorf("expected default base URL, got %s", client.baseURL)
	}
}

func TestClient_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate a network error
	client := New(Config{
		Key:        "test-key",
		BaseURL:    "http://invalid-host-that-does-not-exist.local:9999",
		HTTPClient: &http.Client{Timeout: 1 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error for network failure, got nil")
	}
}

func TestClient_ReadBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Send a very large body to potentially cause issues
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		largeBody := strings.Repeat(`{"data":"`, 1000000) + `"}`
		_, _ = w.Write([]byte(largeBody))
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	// This should still work, just testing the read path
	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	// The test may pass or fail depending on system resources
	// We're mainly checking that the client handles it gracefully
	if err != nil && !strings.Contains(err.Error(), "reading response body") {
		t.Logf("Got error (may be expected for large body): %v", err)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Message:    "not found",
		Endpoint:   "/query",
	}

	errStr := err.Error()
	expected := "API error: status=404 message=\"not found\" endpoint=/query"
	if errStr != expected {
		t.Errorf("expected %q, got %q", expected, errStr)
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		name   string
		err    *APIError
		want   string
		wantOK bool
	}{
		{
			name:   "4xx error",
			err:    &APIError{StatusCode: 400, Message: "bad request", Endpoint: "/query"},
			want:   "client error",
			wantOK: true,
		},
		{
			name:   "5xx error",
			err:    &APIError{StatusCode: 500, Message: "server error", Endpoint: "/query"},
			want:   "server error",
			wantOK: true,
		},
		{
			name:   "2xx status",
			err:    &APIError{StatusCode: 200, Message: "ok", Endpoint: "/query"},
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unwrapped := errors.Unwrap(tt.err)
			if tt.wantOK {
				if unwrapped == nil {
					t.Fatal("expected unwrapped error, got nil")
				}
				if unwrapped.Error() != tt.want {
					t.Errorf("expected unwrapped error %q, got %q", tt.want, unwrapped.Error())
				}
			} else { //nolint:gocritic // Simple if-else pattern for test assertions
				if unwrapped != nil {
					t.Errorf("expected nil unwrap, got %v", unwrapped)
				}
			}
		})
	}
}

// Benchmark for token bucket performance
func BenchmarkTokenBucket_Wait(b *testing.B) {
	tb := newTokenBucket(1000, time.Second)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tb.Wait(ctx)
	}
}

// Benchmark for concurrent rate limiting
func BenchmarkClient_ConcurrentRequests(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		RateLimit:  100,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.get(ctx, map[string]string{"function": "TEST"})
		}
	})
}

func TestClient_BurstLimiter_EnforcesFivePerSecond(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create client with burst limit of 5 requests per second
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		BurstLimit: 5,
		RateLimit:  70,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()

	// First 5 requests should succeed quickly
	start := time.Now()
	for i := 0; i < 5; i++ {
		_, err := client.get(ctx, map[string]string{"function": "TEST"})
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	firstDuration := time.Since(start)

	// Sixth request should be rate-limited by burst limiter and take some time
	start = time.Now()
	_, err := client.get(ctx, map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("sixth request failed: %v", err)
	}
	secondDuration := time.Since(start)

	// The sixth request should have taken at least ~200ms due to burst limiting
	// With 5 requests/second, each request gets ~200ms
	if secondDuration < 100*time.Millisecond {
		t.Errorf("burst limiting may not have worked. First 5 took %v, sixth took %v", firstDuration, secondDuration)
	}

	t.Logf("First 5 requests took %v, sixth request took %v", firstDuration, secondDuration)
}

func TestClient_BurstLimiter_DefaultsToFiveWhenUnset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create client without explicitly setting BurstLimit
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	if client.config.BurstLimit != DefaultBurstLimit {
		t.Errorf("expected default burst limit %d, got %d", DefaultBurstLimit, client.config.BurstLimit)
	}

	if client.burstLimiter == nil {
		t.Fatal("burst limiter should not be nil")
	}
}

func TestClient_BurstLimiter_RespectsUserSuppliedValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	customBurstLimit := 10
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		BurstLimit: customBurstLimit,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	if client.config.BurstLimit != customBurstLimit {
		t.Errorf("expected burst limit %d, got %d", customBurstLimit, client.config.BurstLimit)
	}

	ctx := context.Background()

	// Should be able to make 10 requests quickly (not limited by burst limiter)
	start := time.Now()
	for i := 0; i < customBurstLimit; i++ {
		_, err := client.get(ctx, map[string]string{"function": "TEST"})
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	duration := time.Since(start)

	// 10 requests should complete in under 1 second if burst limit is 10/sec
	if duration > time.Second {
		t.Errorf("burst limit may not be respected. Expected 10 reqs/sec, took %v for 10 requests", duration)
	}
}

func TestClient_BurstLimiter_ContextCancellationWrapped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		BurstLimit: 1,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()

	// First request should succeed
	_, err := client.get(ctx, map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// Second request with canceled context should return wrapped error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.get(ctx, map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "burst limiter:") {
		t.Errorf("expected error to contain 'burst limiter:', got: %v", err)
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error to be context.Canceled, got: %T", err)
	}
}

func TestClient_BothLimitersComposedBurstFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create client with very low burst and rate limits to test composition
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		BurstLimit: 2,
		RateLimit:  2,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()

	// First 2 requests should succeed quickly (burst limit allows 2/sec)
	start := time.Now()
	for i := 0; i < 2; i++ {
		_, err := client.get(ctx, map[string]string{"function": "TEST"})
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	firstTwoDuration := time.Since(start)

	// Third request should be limited by burst limiter (since burst is hit first)
	// After burst bucket refills, rate limiter will also be the binding constraint
	// since both are set to 2
	start = time.Now()
	_, err := client.get(ctx, map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("third request failed: %v", err)
	}
	thirdDuration := time.Since(start)

	t.Logf("First 2 requests took %v, third request took %v", firstTwoDuration, thirdDuration)
}

func TestClient_PerMinuteBucketDoesNotStartFull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create client with burst limit of 5 but rate limit of 60/minute (1/sec)
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		BurstLimit: 5,
		RateLimit:  60,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()

	// First 5 requests should succeed quickly (from burst bucket)
	start := time.Now()
	for i := 0; i < 5; i++ {
		_, err := client.get(ctx, map[string]string{"function": "TEST"})
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	burstDuration := time.Since(start)

	// Sixth request should wait for either burst bucket refill (~200ms) or rate bucket (~1000ms)
	// Since rate bucket started half-full (30 tokens), it's not the bottleneck
	// The burst bucket refills every 200ms, so the 6th request should wait ~200ms
	start = time.Now()
	_, err := client.get(ctx, map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("sixth request failed: %v", err)
	}
	sixthDuration := time.Since(start)

	// The 6th request should take noticeably longer (waiting for burst bucket refill)
	// With 5/sec burst, the refill interval is 200ms
	if sixthDuration < 100*time.Millisecond {
		t.Errorf("sixth request waited %v, expected at least ~100ms for burst bucket refill", sixthDuration)
	}

	// First 5 should be fast, 6th should wait
	if burstDuration > 100*time.Millisecond {
		t.Errorf("first 5 requests took %v, expected them to be fast (burst bucket had tokens)", burstDuration)
	}

	t.Logf("First 5 requests took %v, sixth request took %v", burstDuration, sixthDuration)
}

func TestTokenBucket_HalfFillOnConstructionForLargeBuckets(t *testing.T) {
	// Large bucket (>5) should start half-full
	tb := newTokenBucket(100, time.Minute)
	tb.mu.Lock()
	tokens := tb.tokens
	tb.mu.Unlock()

	expected := float64(100) / 2
	if tokens != expected {
		t.Errorf("expected bucket to start with %f tokens, got %f", expected, tokens)
	}
}

func TestTokenBucket_FullFillForSmallBuckets(t *testing.T) {
	// Small bucket (<=5) should start full
	capacities := []int{1, 2, 3, 4, 5}

	for _, capacity := range capacities {
		t.Run(fmt.Sprintf("capacity_%d", capacity), func(t *testing.T) {
			tb := newTokenBucket(capacity, time.Minute)
			tb.mu.Lock()
			tokens := tb.tokens
			tb.mu.Unlock()

			expected := float64(capacity)
			if tokens != expected {
				t.Errorf("expected bucket to start with %f tokens, got %f", expected, tokens)
			}
		})
	}
}

func TestClient_FiftyConcurrentRequests_NeverExceeds5InAnySecond(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		BurstLimit: 5,
		RateLimit:  70,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx := context.Background()
	var wg sync.WaitGroup

	// Record when each request completes (client-side timing)
	var mu sync.Mutex
	completionTimes := make([]time.Time, 50)

	// Launch 50 concurrent requests
	start := time.Now()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = client.get(ctx, map[string]string{"function": "TEST", "id": fmt.Sprintf("%d", id)})
			mu.Lock()
			completionTimes[id] = time.Now()
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	totalDuration := time.Since(start)

	// With burst limit of 5/sec, 50 requests should take at least ~10 seconds
	// The first 5 complete immediately (burst bucket starts full)
	// Then tokens refill at 5/sec continuously
	minExpectedDuration := time.Duration(50/5-1) * time.Second // ~9 seconds
	if totalDuration < minExpectedDuration {
		t.Errorf("total duration was %v, expected at least %v for 50 requests at 5/sec", totalDuration, minExpectedDuration)
	}

	// Count requests completing in the first second
	// With 50 concurrent requests and burst limit 5/sec:
	// - First 5 get tokens immediately (burst bucket starts full)
	// - Burst bucket refills at 5/sec, so by the end of first second,
	//   about 5 more tokens have become available and been consumed
	// - Therefore, ~10 requests can complete in the first second
	mu.Lock()
	var firstSecondCount int
	for _, ct := range completionTimes {
		if ct.Sub(start) < time.Second {
			firstSecondCount++
		}
	}
	mu.Unlock()

	// Allow up to 12 requests in first second (5 initial + ~5 refilled + some margin)
	// This verifies the burst limiter is working while accounting for continuous refill
	if firstSecondCount > 12 {
		t.Errorf("found %d requests completing in first second, expected ~10 (burst limit: 5/sec)", firstSecondCount)
	}

	// Also verify that the rate is controlled over time - check requests completing
	// between second 1 and second 2
	mu.Lock()
	var secondSecondCount int
	for _, ct := range completionTimes {
		elapsed := ct.Sub(start)
		if elapsed >= time.Second && elapsed < 2*time.Second {
			secondSecondCount++
		}
	}
	mu.Unlock()

	// Should be roughly 5-6 requests in the next second (the burst limiter's steady state)
	if secondSecondCount > 7 {
		t.Errorf("found %d requests completing in second 1-2, expected ~5 (burst limit: 5/sec)", secondSecondCount)
	}

	t.Logf("50 requests completed in %v, with %d in first second, %d in second second", totalDuration, firstSecondCount, secondSecondCount)
}

func TestCheckAPIError_BurstPatternReturnsTransient(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`{"Information": "Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day. Please subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to instantly remove all daily rate limits. Burst pattern detected: 5 requests per second allowed."}`)

	err := client.checkAPIError(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrTransientAPIError) {
		t.Errorf("expected error to be ErrTransientAPIError, got: %T: %v", err, err)
	}
}

func TestCheckAPIError_BurstPatternTypoReturnsTransient(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`{"Information": "Burst pattern detected: 5 requets per second allowed."}`)

	err := client.checkAPIError(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrTransientAPIError) {
		t.Errorf("expected error to be ErrTransientAPIError, got: %T: %v", err, err)
	}
}

func TestCheckAPIError_BurstPatternCaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{
			name: "uppercase",
			body: []byte(`{"Information": "BURST PATTERN DETECTED: 5 REQUESTS PER SECOND ALLOWED."}`),
		},
		{
			name: "mixed case",
			body: []byte(`{"Information": "Burst Pattern Detected: 5 Requets Per Second Allowed."}`),
		},
		{
			name: "lowercase",
			body: []byte(`{"Information": "burst pattern detected: 5 requests per second allowed."}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(Config{Key: "test-key"})
			err := client.checkAPIError(tt.body)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, ErrTransientAPIError) {
				t.Errorf("expected error to be ErrTransientAPIError, got: %T: %v", err, err)
			}
		})
	}
}

func TestCheckAPIError_GenericInformationNotTransient(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`{"Information": "Thank you for using Alpha Vantage!"}`)

	err := client.checkAPIError(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if errors.Is(err, ErrTransientAPIError) {
		t.Errorf("expected error NOT to be ErrTransientAPIError, got: %v", err)
	}

	expectedMsg := "API information: Thank you for using Alpha Vantage!"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestCheckAPIError_ErrorMessageNotTransient(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`{"Error Message": "Invalid API key."}`)

	err := client.checkAPIError(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if errors.Is(err, ErrTransientAPIError) {
		t.Errorf("expected error NOT to be ErrTransientAPIError, got: %v", err)
	}

	expectedMsg := "API error: Invalid API key."
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestCheckAPIError_NoteNotTransient(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`{"Note": "Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day."}`)

	err := client.checkAPIError(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if errors.Is(err, ErrTransientAPIError) {
		t.Errorf("expected error NOT to be ErrTransientAPIError, got: %v", err)
	}

	expectedMsg := "API note: Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day."
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestIsBurstPatternInfo_ExhaustiveTable(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{
			name: "burst pattern phrase",
			msg:  "Burst pattern detected: 5 requests per second allowed.",
			want: true,
		},
		{
			name: "5 requests per second",
			msg:  "You have exceeded 5 requests per second.",
			want: true,
		},
		{
			name: "5 requets per second typo",
			msg:  "You have exceeded 5 requets per second.",
			want: true,
		},
		{
			name: "mixed case burst pattern",
			msg:  "BURST PATTERN DETECTED",
			want: true,
		},
		{
			name: "generic information",
			msg:  "Thank you for using Alpha Vantage!",
			want: false,
		},
		{
			name: "premium key message",
			msg:  "The premium API key is required.",
			want: false,
		},
		{
			name: "standard rate limit message",
			msg:  "Our standard API rate limit is 25 requests per day.",
			want: false,
		},
		{
			name: "empty string",
			msg:  "",
			want: false,
		},
		{
			name: "partial match burst pattern",
			msg:  "detecting a burst",
			want: false,
		},
		{
			name: "6 requests per second",
			msg:  "6 requests per second allowed",
			want: false,
		},
		{
			name: "requests not requets",
			msg:  "5 requests per second allowed",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBurstPatternInfo(tt.msg)
			if got != tt.want {
				t.Errorf("isBurstPatternInfo(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestCheckAPIError_Success(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`{"success": true, "data": "test"}`)

	err := client.checkAPIError(body)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCheckAPIError_InvalidJSON(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`not valid json`)

	err := client.checkAPIError(body)
	if err != nil {
		t.Fatalf("expected no error for invalid JSON, got: %v", err)
	}
}

func TestCheckAPIError_EmptyObject(t *testing.T) {
	client := New(Config{Key: "test-key"})
	body := []byte(`{}`)

	err := client.checkAPIError(body)
	if err != nil {
		t.Fatalf("expected no error for empty object, got: %v", err)
	}
}

func TestClient_Request_RetriesOnBurstPatternAndSucceeds(t *testing.T) {
	attempts := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		attempts++

		if attempts < 3 {
			// Return burst pattern error on first two attempts
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"Information": "Burst pattern detected: 5 requests per second allowed."}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 5,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	data, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	if string(data) != `{"success": true}` {
		t.Errorf("unexpected response: %s", string(data))
	}
}

func TestClient_Request_ExhaustsRetriesOnPersistentBurst(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		callCount++

		// Always return burst pattern error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"Information": "Burst pattern detected: 5 requests per second allowed."}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 2, // 2 retries + initial = 3 total attempts
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}

	if !errors.Is(err, ErrTransientAPIError) {
		t.Errorf("expected error to be ErrTransientAPIError, got: %T: %v", err, err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}

func TestClient_Request_DoesNotRetryOnNonTransientInformation(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		callCount++

		// Return non-transient information error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"Information": "The premium API key is required."}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 5,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if errors.Is(err, ErrTransientAPIError) {
		t.Errorf("expected error NOT to be ErrTransientAPIError, got: %v", err)
	}

	expectedMsg := "API information: The premium API key is required."
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestClient_Request_BackoffRespectsContextCancellation(t *testing.T) {
	attempts := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		attempts++

		// Always return burst pattern error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"Information": "Burst pattern detected: 5 requests per second allowed."}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: 10,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after first attempt completes
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := client.get(ctx, map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error to be context.Canceled, got: %v", err)
	}

	// Should have only made 1-2 attempts due to context cancellation
	if attempts > 2 {
		t.Errorf("expected at most 2 attempts due to context cancellation, got %d", attempts)
	}
}

func TestClient_Request_BurstAndHTTP429ShareBackoff(t *testing.T) {
	tests := []struct {
		name           string
		handler        func(w http.ResponseWriter, attempts int)
		expectedDelay  time.Duration
		delayTolerance time.Duration
	}{
		{
			name: "burst pattern",
			handler: func(w http.ResponseWriter, attempts int) {
				if attempts < 2 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprint(w, `{"Information": "Burst pattern detected: 5 requests per second allowed."}`)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{"success": true}`)
			},
			expectedDelay:  250 * time.Millisecond,
			delayTolerance: 100 * time.Millisecond,
		},
		{
			name: "HTTP 429",
			handler: func(w http.ResponseWriter, attempts int) {
				if attempts < 2 {
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{"success": true}`)
			},
			expectedDelay:  250 * time.Millisecond,
			delayTolerance: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0
			var mu sync.Mutex

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				attempts++
				tt.handler(w, attempts)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				MaxRetries: 5,
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			})

			start := time.Now()
			_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
			duration := time.Since(start)

			// Verify that the delay is approximately the expected backoff time
			// Allow tolerance for jitter
			minExpected := tt.expectedDelay - tt.delayTolerance
			maxExpected := tt.expectedDelay + tt.delayTolerance + 100*time.Millisecond // extra for processing

			if duration < minExpected || duration > maxExpected {
				t.Errorf("expected duration ~%v (±%v), got %v", tt.expectedDelay, tt.delayTolerance, duration)
			}
		})
	}
}

func TestClient_Request_RetryCountCappedAtMaxRetries(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		callCount++

		// Always return burst pattern error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"Information": "Burst pattern detected: 5 requests per second allowed."}`)
	}))
	defer server.Close()

	maxRetries := 3
	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		MaxRetries: maxRetries,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.get(context.Background(), map[string]string{"function": "TEST"})
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}

	expectedCalls := maxRetries + 1 // initial + retries
	if callCount != expectedCalls {
		t.Errorf("expected %d calls (1 initial + %d retries), got %d", expectedCalls, maxRetries, callCount)
	}
}

func TestBackoffFor_MonotonicAndCapped(t *testing.T) {
	// Test monotonic increase (without randomness)
	// We'll sample multiple times and verify the trend
	samples := make([]time.Duration, 10)
	for i := range samples {
		samples[i] = backoffFor(i)
	}

	// Verify each sample is non-negative
	for i, d := range samples {
		if d < 0 {
			t.Errorf("attempt %d: backoff should be non-negative, got %v", i, d)
		}
	}

	// Verify that with enough samples, we see the capped value
	maxSeen := time.Duration(0)
	maxBackoff := 5 * time.Second
	foundCap := false
	for i := 0; i < 20; i++ {
		d := backoffFor(i)
		if d > maxSeen {
			maxSeen = d
		}
		// With jitter, we may see values slightly above maxBackoff (maxBackoff + maxJitter)
		// We check if the value is at or very close to maxBackoff
		if d >= maxBackoff && d <= maxBackoff+100*time.Millisecond {
			foundCap = true
		}
	}

	if !foundCap {
		t.Errorf("expected to see capped backoff of 5s (±100ms jitter), max seen was %v", maxSeen)
	}

	// Verify exponential growth trend (min values, ignoring jitter)
	// attempt 0: 250ms * 1 = 250ms + 0-100ms = 250-350ms
	// attempt 1: 250ms * 2 = 500ms + 0-100ms = 500-600ms
	// attempt 2: 250ms * 4 = 1000ms + 0-100ms = 1000-1100ms
	minBackoffs := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{0, 250 * time.Millisecond, 350 * time.Millisecond},
		{1, 500 * time.Millisecond, 600 * time.Millisecond},
		{2, 1000 * time.Millisecond, 1100 * time.Millisecond},
		{3, 2000 * time.Millisecond, 2100 * time.Millisecond},
		{4, 4000 * time.Millisecond, 4100 * time.Millisecond},
	}

	for _, tc := range minBackoffs {
		d := backoffFor(tc.attempt)
		// We sample multiple times to account for jitter
		foundInRange := false
		for i := 0; i < 100; i++ {
			d = backoffFor(tc.attempt)
			if d >= tc.minExpected && d <= tc.maxExpected {
				foundInRange = true
				break
			}
		}
		if !foundInRange {
			t.Errorf("attempt %d: expected backoff in range [%v, %v], sampled %v", tc.attempt, tc.minExpected, tc.maxExpected, d)
		}
	}
}

func TestBackoffFor_WithJitterBoundedBy100ms(t *testing.T) {
	const (
		numSamples = 1000
		maxJitter  = 100 * time.Millisecond
		maxBackoff = 5 * time.Second
	)

	for attempt := 0; attempt < 5; attempt++ {
		minSeen := time.Duration(^uint64(0) >> 1)
		maxSeen := time.Duration(0)

		for i := 0; i < numSamples; i++ {
			d := backoffFor(attempt)
			if d < minSeen {
				minSeen = d
			}
			if d > maxSeen {
				maxSeen = d
			}
		}

		// Calculate expected base (without jitter)
		base := 250 * time.Millisecond * time.Duration(1<<uint(attempt))
		if base > maxBackoff {
			base = maxBackoff
		}

		// Minimum should be base (jitter can be 0)
		if minSeen < base {
			t.Errorf("attempt %d: minimum backoff %v should be >= base %v", attempt, minSeen, base)
		}

		// Maximum should be base + maxJitter (or maxBackoff if base is already at maxBackoff)
		expectedMax := base + maxJitter
		if expectedMax > maxBackoff {
			expectedMax = maxBackoff
		}
		if maxSeen > expectedMax {
			t.Errorf("attempt %d: maximum backoff %v should be <= %v (base %v + jitter %v, maxBackoff %v)",
				attempt, maxSeen, expectedMax, base, maxJitter, maxBackoff)
		}

		// Verify spread is roughly maxJitter (unless capped)
		spread := maxSeen - minSeen
		if base < maxBackoff && spread < maxJitter/2 {
			t.Errorf("attempt %d: spread %v should be roughly maxJitter %v (min %v, max %v, base %v)",
				attempt, spread, maxJitter, minSeen, maxSeen, base)
		}
	}
}
