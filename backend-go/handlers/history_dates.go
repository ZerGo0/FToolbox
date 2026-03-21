package handlers

import "time"

func parseHistoryDate(value string) *time.Time {
	if value == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return &parsed
	}

	parsed, err = time.Parse("2006-01-02", value)
	if err == nil {
		return &parsed
	}

	return nil
}
