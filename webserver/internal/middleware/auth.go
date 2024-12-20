package middleware

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AuthConfig contains authentication configuration
type AuthConfig struct {
	APIKey      string
	TokenHeader string
	SkipPaths   []string
	TokenExpiry time.Duration
	RateLimit   int // Requests per minute
	CacheSize   int // Size of token cache
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		TokenHeader: "Authorization",
		TokenExpiry: 24 * time.Hour,
		RateLimit:   60,
		CacheSize:   1000,
	}
}

// Auth creates a new authentication middleware
func Auth(apiKey string) func(http.Handler) http.Handler {
	cfg := DefaultAuthConfig()
	cfg.APIKey = apiKey

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			if shouldSkipAuth(r.URL.Path, cfg.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from header
			token := extractToken(r.Header.Get(cfg.TokenHeader))
			if token == "" {
				http.Error(w, "Missing authentication token", http.StatusUnauthorized)
				return
			}

			// Validate token
			if !validateToken(token, cfg.APIKey) {
				http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
				return
			}

			// Add authentication info to context
			ctx := context.WithValue(r.Context(), "auth_info", AuthInfo{
				Token:    token,
				IssuedAt: time.Now(),
			})

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthInfo contains authentication information
type AuthInfo struct {
	Token     string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// extractToken extracts the token from the Authorization header
func extractToken(header string) string {
	if header == "" {
		return ""
	}

	// Check for Bearer token
	parts := strings.Split(header, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}

	// Return header as-is for API key
	return header
}

// validateToken validates the authentication token
func validateToken(token, apiKey string) bool {
	// Use constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(token), []byte(apiKey)) == 1
}

// shouldSkipAuth checks if authentication should be skipped for the path
func shouldSkipAuth(path string, skipPaths []string) bool {
	// Skip health check and metrics endpoints
	if path == "/health" || path == "/metrics" {
		return true
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// Chain combines multiple middleware handlers
func Chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// RateLimit creates a new rate limiting middleware
func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	type client struct {
		count    int
		lastSeen time.Time
	}
	clients := make(map[string]*client)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = strings.Split(forwarded, ",")[0]
			}

			// Check rate limit
			now := time.Now()
			if c, exists := clients[clientIP]; exists {
				// Reset count if minute has passed
				if now.Sub(c.lastSeen) >= time.Minute {
					c.count = 0
					c.lastSeen = now
				}

				// Check if rate limit exceeded
				if c.count >= requestsPerMinute {
					w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", c.lastSeen.Add(time.Minute).Unix()))
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}

				c.count++
			} else {
				clients[clientIP] = &client{
					count:    1,
					lastSeen: now,
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
