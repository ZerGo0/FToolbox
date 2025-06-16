package ratelimit

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// GlobalRateLimiter enforces a global rate limit across all endpoints
type GlobalRateLimiter struct {
	mu                sync.Mutex
	maxRequests       int
	window            time.Duration
	requestTimestamps []time.Time
	logger            *zap.Logger
}

// NewGlobalRateLimiter creates a new global rate limiter
func NewGlobalRateLimiter(maxRequests int, windowSeconds int, logger *zap.Logger) *GlobalRateLimiter {
	return &GlobalRateLimiter{
		maxRequests:       maxRequests,
		window:            time.Duration(windowSeconds) * time.Second,
		requestTimestamps: make([]time.Time, 0),
		logger:            logger,
	}
}

// Wait blocks until a request can be made within the global rate limit
func (g *GlobalRateLimiter) Wait(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-g.window)

	// Clean old timestamps
	validTimestamps := make([]time.Time, 0, len(g.requestTimestamps))
	for _, ts := range g.requestTimestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	g.requestTimestamps = validTimestamps

	// Check if we're at the limit
	if len(g.requestTimestamps) >= g.maxRequests {
		// Calculate wait time - we need to wait until the oldest request expires
		oldestRequest := g.requestTimestamps[0]
		waitTime := oldestRequest.Add(g.window).Sub(now)

		if waitTime > 0 {
			g.logger.Info("Global rate limit reached, waiting",
				zap.Int("current_requests", len(g.requestTimestamps)),
				zap.Int("max_requests", g.maxRequests),
				zap.Duration("window", g.window),
				zap.Duration("wait_time", waitTime))

			g.mu.Unlock()
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return ctx.Err()
			}
			g.mu.Lock()

			// Re-clean after waiting
			now = time.Now()
			cutoff = now.Add(-g.window)
			validTimestamps = validTimestamps[:0]
			for _, ts := range g.requestTimestamps {
				if ts.After(cutoff) {
					validTimestamps = append(validTimestamps, ts)
				}
			}
			g.requestTimestamps = validTimestamps
		}
	}

	// Record this request
	g.requestTimestamps = append(g.requestTimestamps, now)

	// Calculate delay to spread requests evenly
	if g.maxRequests > 1 && len(g.requestTimestamps) > 1 {
		idealSpacing := g.window / time.Duration(g.maxRequests)
		lastRequest := g.requestTimestamps[len(g.requestTimestamps)-2]
		timeSinceLastRequest := now.Sub(lastRequest)

		if timeSinceLastRequest < idealSpacing {
			spreadDelay := idealSpacing - timeSinceLastRequest
			g.logger.Debug("Spreading requests",
				zap.Duration("ideal_spacing", idealSpacing),
				zap.Duration("actual_spacing", timeSinceLastRequest),
				zap.Duration("delay", spreadDelay))

			g.mu.Unlock()
			select {
			case <-time.After(spreadDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
			g.mu.Lock()
		}
	}

	return nil
}

// GetStats returns current statistics
func (g *GlobalRateLimiter) GetStats() map[string]interface{} {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-g.window)

	activeRequests := 0
	for _, ts := range g.requestTimestamps {
		if ts.After(cutoff) {
			activeRequests++
		}
	}

	return map[string]interface{}{
		"max_requests":    g.maxRequests,
		"window":          g.window.String(),
		"active_requests": activeRequests,
		"capacity_used":   float64(activeRequests) / float64(g.maxRequests) * 100,
	}
}
