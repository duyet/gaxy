package proxy

import (
	"fmt"
	"net/url"
	"strings"
)

// sanitizeRequestURI validates and sanitizes a request URI to prevent SSRF attacks.
// It ensures the URI is a safe path+query combination, not a full URL.
func sanitizeRequestURI(reqURI string) (string, error) {
	// Reject empty URIs
	if reqURI == "" {
		return "", fmt.Errorf("request URI cannot be empty")
	}

	// Reject URIs that look like full URLs (contain scheme)
	if strings.Contains(reqURI, "://") {
		return "", fmt.Errorf("request URI must not contain a scheme (http://)")
	}

	// Reject URIs that start with // (protocol-relative URLs)
	if strings.HasPrefix(reqURI, "//") {
		return "", fmt.Errorf("request URI must not be a protocol-relative URL")
	}

	// Parse the URI to validate its structure
	parsedURI, err := url.Parse(reqURI)
	if err != nil {
		return "", fmt.Errorf("invalid request URI: %w", err)
	}

	// Ensure there's no host component (which would indicate a full URL)
	if parsedURI.Host != "" {
		return "", fmt.Errorf("request URI must not contain a host")
	}

	// Ensure there's no scheme (extra safety check)
	if parsedURI.Scheme != "" {
		return "", fmt.Errorf("request URI must not contain a scheme")
	}

	// Ensure the path starts with /
	path := parsedURI.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("request URI path must start with /")
	}

	// Prevent directory traversal attempts
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("request URI must not contain directory traversal sequences")
	}

	// Reconstruct a safe URI with just path and query
	safeURI := path
	if parsedURI.RawQuery != "" {
		safeURI += "?" + parsedURI.RawQuery
	}

	return safeURI, nil
}

// isAllowedPath checks if a path is allowed for proxying to Google Analytics.
// This provides an additional security layer by whitelisting known GA/GTM paths.
func isAllowedPath(path string) bool {
	// Common Google Analytics and Tag Manager endpoints
	allowedPrefixes := []string{
		"/analytics.js",
		"/ga.js",
		"/gtag/js",
		"/gtm.js",
		"/collect",
		"/j/collect",
		"/g/collect",
		"/r/collect",
		"/batch",
		"/api/",
	}

	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
