package ratelimit

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EndpointConfig holds rate limit configuration for a specific endpoint
type EndpointConfig struct {
	// Current rate limit (requests per window)
	Limit int
	// Time window for the limit
	Window time.Duration
	// Last time we hit a rate limit
	LastRateLimitHit time.Time
	// Number of consecutive successful requests
	SuccessStreak int
	// Number of rate limit hits
	RateLimitHits int
	// Backoff multiplier for retries
	BackoffMultiplier float64
	// Current backoff duration
	CurrentBackoff time.Duration
	// Request timestamps for sliding window
	RequestTimestamps []time.Time
}

// AdaptiveRateLimiter manages rate limits per endpoint with learning capability
type AdaptiveRateLimiter struct {
	mu        sync.RWMutex
	endpoints map[string]*EndpointConfig
	logger    *zap.Logger

	// Default configuration for new endpoints
	defaultLimit      int
	defaultWindow     time.Duration
	maxBackoff        time.Duration
	minBackoff        time.Duration
	backoffMultiplier float64

	// Persistence
	persistFunc func(map[string]*EndpointConfig) error
	loadFunc    func() (map[string]*EndpointConfig, error)

	// Global rate limiter
	globalLimiter *GlobalRateLimiter
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter(logger *zap.Logger, defaultLimit int) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		endpoints:         make(map[string]*EndpointConfig),
		logger:            logger,
		defaultLimit:      defaultLimit,
		defaultWindow:     time.Minute,
		maxBackoff:        5 * time.Minute,
		minBackoff:        time.Second,
		backoffMultiplier: 2.0,
	}
}

// SetGlobalLimiter sets the global rate limiter
func (r *AdaptiveRateLimiter) SetGlobalLimiter(limiter *GlobalRateLimiter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.globalLimiter = limiter
}

// SetPersistence sets functions for saving and loading learned configurations
func (r *AdaptiveRateLimiter) SetPersistence(save func(map[string]*EndpointConfig) error, load func() (map[string]*EndpointConfig, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.persistFunc = save
	r.loadFunc = load

	// Load existing configurations
	if load != nil {
		configs, err := load()
		if err != nil {
			r.logger.Warn("Failed to load rate limit configurations", zap.Error(err))
		} else if configs != nil {
			r.endpoints = configs
			r.logger.Info("Loaded rate limit configurations", zap.Int("endpoints", len(configs)))
		}
	}

	return nil
}

// WaitIfNeeded blocks until the request can be made without exceeding rate limits
func (r *AdaptiveRateLimiter) WaitIfNeeded(ctx context.Context, endpoint string) error {
	// First check global rate limit if configured
	if r.globalLimiter != nil {
		if err := r.globalLimiter.Wait(ctx); err != nil {
			return err
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	config := r.getOrCreateEndpoint(endpoint)

	// Clean old timestamps
	now := time.Now()
	cutoff := now.Add(-config.Window)
	validTimestamps := make([]time.Time, 0, len(config.RequestTimestamps))
	for _, ts := range config.RequestTimestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	config.RequestTimestamps = validTimestamps

	// Check if we're in backoff period
	if config.CurrentBackoff > 0 && config.LastRateLimitHit.Add(config.CurrentBackoff).After(now) {
		waitTime := config.LastRateLimitHit.Add(config.CurrentBackoff).Sub(now)
		r.logger.Info("In backoff period, waiting",
			zap.String("endpoint", endpoint),
			zap.Duration("wait", waitTime),
			zap.Duration("backoff", config.CurrentBackoff))

		r.mu.Unlock()
		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return ctx.Err()
		}
		r.mu.Lock()
	}

	// Check if we need to wait for rate limit
	if len(config.RequestTimestamps) >= config.Limit {
		oldestRequest := config.RequestTimestamps[0]
		waitTime := oldestRequest.Add(config.Window).Sub(now)
		if waitTime > 0 {
			r.logger.Info("At rate limit, waiting",
				zap.String("endpoint", endpoint),
				zap.Int("current", len(config.RequestTimestamps)),
				zap.Int("limit", config.Limit),
				zap.Duration("wait", waitTime))

			r.mu.Unlock()
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return ctx.Err()
			}
			r.mu.Lock()

			// Re-clean timestamps after waiting
			now = time.Now()
			cutoff = now.Add(-config.Window)
			validTimestamps = validTimestamps[:0]
			for _, ts := range config.RequestTimestamps {
				if ts.After(cutoff) {
					validTimestamps = append(validTimestamps, ts)
				}
			}
			config.RequestTimestamps = validTimestamps
		}
	}

	// Record this request
	config.RequestTimestamps = append(config.RequestTimestamps, now)

	return nil
}

// HandleResponse processes the response and adjusts rate limits based on headers
func (r *AdaptiveRateLimiter) HandleResponse(endpoint string, resp *http.Response) {
	r.mu.Lock()
	defer r.mu.Unlock()

	config := r.getOrCreateEndpoint(endpoint)

	// Handle rate limit error
	if resp.StatusCode == http.StatusTooManyRequests {
		config.LastRateLimitHit = time.Now()
		config.RateLimitHits++
		config.SuccessStreak = 0

		// Increase backoff
		if config.CurrentBackoff == 0 {
			config.CurrentBackoff = r.minBackoff
		} else {
			config.CurrentBackoff = time.Duration(float64(config.CurrentBackoff) * config.BackoffMultiplier)
			if config.CurrentBackoff > r.maxBackoff {
				config.CurrentBackoff = r.maxBackoff
			}
		}

		// Try to parse retry-after header
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				config.CurrentBackoff = time.Duration(seconds) * time.Second
			}
		}

		// Reduce our internal limit to be more conservative
		config.Limit = int(float64(config.Limit) * 0.8)
		if config.Limit < 1 {
			config.Limit = 1
		}

		r.logger.Warn("Rate limit hit, adjusting",
			zap.String("endpoint", endpoint),
			zap.Int("new_limit", config.Limit),
			zap.Duration("backoff", config.CurrentBackoff),
			zap.Int("total_hits", config.RateLimitHits))

		r.persist()
		return
	}

	// Success - update streak and potentially adjust limits
	config.SuccessStreak++

	// Reset backoff on success
	if config.SuccessStreak > 10 {
		config.CurrentBackoff = 0
	}

	// Parse rate limit headers if available
	r.parseRateLimitHeaders(endpoint, resp)

	// Gradually increase limit if we've been successful for a while
	if config.SuccessStreak > 100 && config.RateLimitHits == 0 {
		config.Limit = int(float64(config.Limit) * 1.1)
		r.logger.Info("Increasing rate limit after success streak",
			zap.String("endpoint", endpoint),
			zap.Int("new_limit", config.Limit),
			zap.Int("streak", config.SuccessStreak))
		config.SuccessStreak = 0 // Reset to avoid continuous increases
		r.persist()
	}
}

// parseRateLimitHeaders extracts rate limit information from response headers
func (r *AdaptiveRateLimiter) parseRateLimitHeaders(endpoint string, resp *http.Response) {
	config := r.endpoints[endpoint]

	// Common rate limit headers
	headers := map[string]string{
		"X-RateLimit-Limit":      resp.Header.Get("X-RateLimit-Limit"),
		"X-Rate-Limit-Limit":     resp.Header.Get("X-Rate-Limit-Limit"),
		"RateLimit-Limit":        resp.Header.Get("RateLimit-Limit"),
		"X-RateLimit-Remaining":  resp.Header.Get("X-RateLimit-Remaining"),
		"X-Rate-Limit-Remaining": resp.Header.Get("X-Rate-Limit-Remaining"),
		"RateLimit-Remaining":    resp.Header.Get("RateLimit-Remaining"),
		"X-RateLimit-Reset":      resp.Header.Get("X-RateLimit-Reset"),
		"X-Rate-Limit-Reset":     resp.Header.Get("X-Rate-Limit-Reset"),
		"RateLimit-Reset":        resp.Header.Get("RateLimit-Reset"),
	}

	// Try to find limit
	for key, value := range headers {
		if value != "" && (key == "X-RateLimit-Limit" || key == "X-Rate-Limit-Limit" || key == "RateLimit-Limit") {
			if limit, err := strconv.Atoi(value); err == nil && limit > 0 {
				if limit != config.Limit {
					r.logger.Info("Discovered actual rate limit from headers",
						zap.String("endpoint", endpoint),
						zap.Int("old_limit", config.Limit),
						zap.Int("new_limit", limit),
						zap.String("header", key))
					config.Limit = int(float64(limit) * 0.9) // Use 90% of actual limit to be safe
					r.persist()
				}
				break
			}
		}
	}
}

// getOrCreateEndpoint returns config for endpoint, creating if needed
func (r *AdaptiveRateLimiter) getOrCreateEndpoint(endpoint string) *EndpointConfig {
	if config, exists := r.endpoints[endpoint]; exists {
		return config
	}

	r.logger.Debug("Creating new endpoint config",
		zap.String("endpoint", endpoint),
		zap.Int("default_limit", r.defaultLimit))

	config := &EndpointConfig{
		Limit:             r.defaultLimit,
		Window:            r.defaultWindow,
		BackoffMultiplier: r.backoffMultiplier,
		RequestTimestamps: make([]time.Time, 0),
	}
	r.endpoints[endpoint] = config
	return config
}

// persist saves current configurations if persistence is configured
func (r *AdaptiveRateLimiter) persist() {
	if r.persistFunc != nil {
		if err := r.persistFunc(r.endpoints); err != nil {
			r.logger.Error("Failed to persist rate limit configurations", zap.Error(err))
		}
	}
}

// GetStats returns current statistics for all endpoints
func (r *AdaptiveRateLimiter) GetStats() map[string]map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for endpoint, config := range r.endpoints {
		stats[endpoint] = map[string]interface{}{
			"limit":              config.Limit,
			"window":             config.Window.String(),
			"success_streak":     config.SuccessStreak,
			"rate_limit_hits":    config.RateLimitHits,
			"current_backoff":    config.CurrentBackoff.String(),
			"requests_in_window": len(config.RequestTimestamps),
		}
	}

	// Add global limiter stats if available
	if r.globalLimiter != nil {
		stats["_global"] = r.globalLimiter.GetStats()
	}

	return stats
}

// Reset clears all learned configurations
func (r *AdaptiveRateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.endpoints = make(map[string]*EndpointConfig)
	r.persist()
}
