package ratelimit

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries        int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
	RetryableStatuses map[int]bool
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        5,
		InitialBackoff:    time.Second,
		MaxBackoff:        time.Minute,
		BackoffMultiplier: 2.0,
		RetryableStatuses: map[int]bool{
			http.StatusTooManyRequests:     true,
			http.StatusInternalServerError: true,
			http.StatusBadGateway:          true,
			http.StatusServiceUnavailable:  true,
			http.StatusGatewayTimeout:      true,
		},
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err        error
	StatusCode int
	RetryAfter time.Duration
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error (status=%d): %v", e.StatusCode, e.Err)
}

// RetryClient wraps HTTP requests with retry logic
type RetryClient struct {
	client      *http.Client
	rateLimiter *AdaptiveRateLimiter
	config      *RetryConfig
	logger      *zap.Logger
}

// NewRetryClient creates a new retry client
func NewRetryClient(client *http.Client, rateLimiter *AdaptiveRateLimiter, logger *zap.Logger) *RetryClient {
	return &RetryClient{
		client:      client,
		rateLimiter: rateLimiter,
		config:      DefaultRetryConfig(),
		logger:      logger,
	}
}

// SetRetryConfig updates the retry configuration
func (c *RetryClient) SetRetryConfig(config *RetryConfig) {
	c.config = config
}

// Do executes an HTTP request with retry logic
func (c *RetryClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	c.logger.Debug("Making request",
		zap.String("endpoint", endpoint),
		zap.String("url", req.URL.String()))

	var lastErr error
	backoff := c.config.InitialBackoff

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Wait for rate limiter
		if err := c.rateLimiter.WaitIfNeeded(ctx, endpoint); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		// Clone request for retry
		reqCopy := req.Clone(ctx)

		// Execute request
		resp, err := c.client.Do(reqCopy)
		if err != nil {
			lastErr = err
			c.logger.Warn("Request failed",
				zap.String("endpoint", endpoint),
				zap.Int("attempt", attempt+1),
				zap.Error(err))

			// Network errors are retryable
			if attempt < c.config.MaxRetries {
				c.sleep(ctx, backoff)
				backoff = c.calculateBackoff(backoff)
				continue
			}
			return nil, err
		}

		// Let rate limiter process the response
		c.rateLimiter.HandleResponse(endpoint, resp)

		// Check if response is retryable
		if c.config.RetryableStatuses[resp.StatusCode] {
			lastErr = &RetryableError{
				Err:        fmt.Errorf("HTTP %d", resp.StatusCode),
				StatusCode: resp.StatusCode,
			}

			// Special handling for rate limit errors
			if resp.StatusCode == http.StatusTooManyRequests {
				// The rate limiter already handles backoff for 429s
				c.logger.Warn("Rate limit error, will retry after rate limiter backoff",
					zap.String("endpoint", endpoint),
					zap.Int("attempt", attempt+1))
			} else {
				// Other retryable errors
				c.logger.Warn("Retryable error",
					zap.String("endpoint", endpoint),
					zap.Int("status", resp.StatusCode),
					zap.Int("attempt", attempt+1))

				if attempt < c.config.MaxRetries {
					resp.Body.Close()
					c.sleep(ctx, backoff)
					backoff = c.calculateBackoff(backoff)
					continue
				}
			}

			// Don't retry 429s here, let the rate limiter handle it
			if resp.StatusCode != http.StatusTooManyRequests && attempt < c.config.MaxRetries {
				resp.Body.Close()
				continue
			}
		}

		// Success or non-retryable error
		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// calculateBackoff calculates the next backoff duration
func (c *RetryClient) calculateBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * c.config.BackoffMultiplier)
	if next > c.config.MaxBackoff {
		next = c.config.MaxBackoff
	}

	// Add jitter (±10%)
	jitter := time.Duration(float64(next) * 0.1 * (2*math.Min(1, math.Max(0, float64(time.Now().UnixNano()%100)/100)) - 1))
	return next + jitter
}

// sleep waits for the specified duration or until context is cancelled
func (c *RetryClient) sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}
}
