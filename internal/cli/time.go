package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseDuration parses human-friendly duration strings like "8 hours ago", "24h", "7 days ago"
func parseDuration(s string) (time.Time, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, " ago")

	// Try standard Go duration first (24h, 30m, etc.)
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d), nil
	}

	// Try ISO 8601 date/datetime
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	// Parse human-friendly formats like "8 hours", "7 days", "2 weeks"
	re := regexp.MustCompile(`^(\d+)\s*(minute|min|m|hour|hr|h|day|d|week|w|month|mon)s?$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return time.Time{}, fmt.Errorf("unable to parse duration: %s", s)
	}

	n, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	var d time.Duration
	switch unit {
	case "minute", "min", "m":
		d = time.Duration(n) * time.Minute
	case "hour", "hr", "h":
		d = time.Duration(n) * time.Hour
	case "day", "d":
		d = time.Duration(n) * 24 * time.Hour
	case "week", "w":
		d = time.Duration(n) * 7 * 24 * time.Hour
	case "month", "mon":
		d = time.Duration(n) * 30 * 24 * time.Hour // Approximate
	default:
		return time.Time{}, fmt.Errorf("unknown time unit: %s", unit)
	}

	return time.Now().Add(-d), nil
}

// parseTimeArg parses a --from or --to argument (ISO 8601 format)
func parseTimeArg(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	// Try various ISO 8601 formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s (use ISO 8601 format like 2006-01-02T15:04:05)", s)
}
