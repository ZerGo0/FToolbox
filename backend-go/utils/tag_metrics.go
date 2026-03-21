package utils

import "strings"

func CalculateRatio(viewCount, postCount int64) float64 {
	if postCount <= 0 {
		return 0
	}

	return float64(viewCount) / float64(postCount)
}

func TagNameHasPlus(tag string) bool {
	return strings.Contains(strings.TrimSpace(tag), "+")
}
