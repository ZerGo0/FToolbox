package utils

import (
	"math"
	"time"
)

// CalculateHeatScore calculates a heat score for a tag based on view count, post count and time
// The algorithm is inspired by Reddit's hot algorithm with modifications
func CalculateHeatScore(viewCount, postCount int64, createdAt, lastCheckedAt time.Time) float64 {
	// If deleted tag (viewCount = 0), return 0 heat
	if viewCount == 0 {
		return 0
	}

	// Calculate engagement score: weighted combination of view count and post count
	// Post count is weighted higher (70%) than view count (30%) since it indicates more active usage
	engagementScore := float64(postCount)*0.7 + float64(viewCount)*0.0003 // Scale down views since they're typically much larger

	// Apply logarithmic scaling to handle large numbers better
	// Adding 1 to avoid log(0) and ensure minimum score
	logScore := math.Log10(engagementScore + 1)

	// Calculate time decay factor
	// Use last checked time for recency, as it indicates when the tag was last active
	hoursSinceLastCheck := time.Since(lastCheckedAt).Hours()

	// Apply exponential decay - score halves every 48 hours
	// This ensures recent activity is valued higher
	decayFactor := math.Pow(0.5, hoursSinceLastCheck/48.0)

	// Calculate final heat score
	heatScore := logScore * decayFactor * 1000 // Scale up for better readability

	// Round to 2 decimal places
	return math.Round(heatScore*100) / 100
}
