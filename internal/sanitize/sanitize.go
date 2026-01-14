// Package sanitize provides utilities for sanitizing user input for safe logging.
package sanitize

import (
	"regexp"
	"strings"
)

// logSanitizer matches potentially dangerous log injection patterns.
var logSanitizer = regexp.MustCompile(`[\r\n\t]`)

// controlCharRemover matches control characters.
var controlCharRemover = regexp.MustCompile(`[\x00-\x1f\x7f]`)

// ForLog sanitizes a string for safe logging by removing newlines
// and other control characters that could be used for log injection.
func ForLog(s string) string {
	// Replace newlines and tabs with escaped versions
	s = logSanitizer.ReplaceAllString(s, " ")
	// Remove other control characters
	s = controlCharRemover.ReplaceAllString(s, "")
	// Truncate very long strings to prevent log flooding
	const maxLogLength = 500
	if len(s) > maxLogLength {
		s = s[:maxLogLength] + "...[truncated]"
	}
	return s
}

// URL sanitizes a URL for safe logging by removing credentials
// and potentially dangerous characters.
func URL(rawURL string) string {
	// Remove potential credentials from URL
	sanitized := rawURL
	if idx := strings.Index(sanitized, "@"); idx != -1 {
		// Find the protocol separator
		if protoIdx := strings.Index(sanitized, "://"); protoIdx != -1 && protoIdx < idx {
			// Replace the credential part with [REDACTED]
			sanitized = sanitized[:protoIdx+3] + "[REDACTED]@" + sanitized[idx+1:]
		}
	}
	return ForLog(sanitized)
}

// Path sanitizes a path for safe logging.
func Path(path string) string {
	// Limit path length and remove control characters
	const maxPathLength = 200
	s := ForLog(path)
	if len(s) > maxPathLength {
		s = s[:maxPathLength] + "...[truncated]"
	}
	return s
}

// Username sanitizes a username for safe logging.
func Username(username string) string {
	// Limit username length and remove control characters
	const maxUsernameLength = 100
	s := ForLog(username)
	if len(s) > maxUsernameLength {
		s = s[:maxUsernameLength] + "...[truncated]"
	}
	return s
}

// Content sanitizes content for safe logging (for notifications, etc.).
func Content(content string) string {
	const maxContentLength = 200
	s := ForLog(content)
	if len(s) > maxContentLength {
		s = s[:maxContentLength] + "...[truncated]"
	}
	return s
}

// CommandOutput sanitizes command output for safe logging.
func CommandOutput(output string) string {
	const maxOutputLength = 1000
	s := ForLog(output)
	if len(s) > maxOutputLength {
		s = s[:maxOutputLength] + "...[truncated]"
	}
	return s
}
