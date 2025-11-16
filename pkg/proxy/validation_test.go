package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeRequestURI(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		shouldErr bool
	}{
		{
			name:      "valid simple path",
			input:     "/analytics.js",
			expected:  "/analytics.js",
			shouldErr: false,
		},
		{
			name:      "valid path with query",
			input:     "/collect?v=1&tid=UA-12345",
			expected:  "/collect?v=1&tid=UA-12345",
			shouldErr: false,
		},
		{
			name:      "empty URI",
			input:     "",
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "full URL with http scheme - ATTACK",
			input:     "http://evil.com/steal",
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "full URL with https scheme - ATTACK",
			input:     "https://internal-service/admin",
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "protocol-relative URL - ATTACK",
			input:     "//evil.com/steal",
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "directory traversal - ATTACK",
			input:     "/../../etc/passwd",
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "path with encoded characters",
			input:     "/collect?v=1&dl=http%3A%2F%2Fexample.com",
			expected:  "/collect?v=1&dl=http%3A%2F%2Fexample.com",
			shouldErr: false,
		},
		{
			name:      "valid GTM path",
			input:     "/gtag/js?id=G-12345",
			expected:  "/gtag/js?id=G-12345",
			shouldErr: false,
		},
		{
			name:      "missing leading slash",
			input:     "collect?v=1",
			expected:  "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizeRequestURI(tt.input)

			if tt.shouldErr {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				assert.NoError(t, err, "Unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "Sanitized URI mismatch")
			}
		})
	}
}

func TestIsAllowedPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "analytics.js - allowed",
			path:     "/analytics.js",
			expected: true,
		},
		{
			name:     "ga.js - allowed",
			path:     "/ga.js",
			expected: true,
		},
		{
			name:     "gtag/js - allowed",
			path:     "/gtag/js",
			expected: true,
		},
		{
			name:     "gtm.js - allowed",
			path:     "/gtm.js",
			expected: true,
		},
		{
			name:     "collect endpoint - allowed",
			path:     "/collect",
			expected: true,
		},
		{
			name:     "batch endpoint - allowed",
			path:     "/batch",
			expected: true,
		},
		{
			name:     "api endpoint - allowed",
			path:     "/api/debug",
			expected: true,
		},
		{
			name:     "random path - not allowed",
			path:     "/random/path",
			expected: false,
		},
		{
			name:     "admin path - not allowed",
			path:     "/admin",
			expected: false,
		},
		{
			name:     "root path - not allowed",
			path:     "/",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAllowedPath(tt.path)
			assert.Equal(t, tt.expected, result, "Path allowance mismatch for: %s", tt.path)
		})
	}
}

// TestSSRFPrevention tests various SSRF attack vectors
func TestSSRFPrevention(t *testing.T) {
	attackVectors := []string{
		"http://169.254.169.254/latest/meta-data/",  // AWS metadata
		"http://metadata.google.internal/",          // GCP metadata
		"http://127.0.0.1:8080/admin",               // Localhost access
		"http://internal-service/secrets",           // Internal service
		"//attacker.com/steal",                      // Protocol-relative
		"https://evil.com/data",                     // External HTTPS
		"file:///etc/passwd",                        // File scheme
		"ftp://internal-ftp/files",                  // FTP scheme
	}

	for _, attack := range attackVectors {
		t.Run("SSRF_Attack_"+attack, func(t *testing.T) {
			result, err := sanitizeRequestURI(attack)
			assert.Error(t, err, "SSRF attack should be blocked: %s", attack)
			assert.Empty(t, result, "Result should be empty for blocked attack")
		})
	}
}
