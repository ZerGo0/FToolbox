package models

import (
	"time"
)

// RateLimit stores learned rate limit configurations for API endpoints
type RateLimit struct {
	ID                string    `gorm:"primaryKey;size:255" json:"id"` // Endpoint identifier
	Limit             int       `json:"limit"`                         // Requests per window
	Window            int64     `json:"window"`                        // Window duration in seconds
	SuccessStreak     int       `json:"success_streak"`
	RateLimitHits     int       `json:"rate_limit_hits"`
	BackoffMultiplier float64   `json:"backoff_multiplier"`
	LastRateLimitHit  time.Time `json:"last_rate_limit_hit"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
