package logutil

import "strings"

// SQLPreview returns a single-line, length-limited preview safe for debug logs.
func SQLPreview(sql string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 120
	}
	s := strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
	if len(s) <= maxLen {
		return s
	}
	return strings.TrimSpace(s[:maxLen]) + "..."
}
