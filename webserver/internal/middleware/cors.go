package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig contains CORS configuration options
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS creates a new CORS middleware handler
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	cfg := DefaultCORSConfig()
	if len(allowedOrigins) > 0 {
		cfg.AllowedOrigins = allowedOrigins
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Skip if no Origin header
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range cfg.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Origin not allowed", http.StatusForbidden)
				return
			}

			// Set CORS headers
			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			headers.Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			headers.Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
			headers.Set("Access-Control-Max-Age", string(cfg.MaxAge))

			if cfg.AllowCredentials {
				headers.Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AllowOrigin checks if the origin is allowed
func (c *CORSConfig) AllowOrigin(origin string) bool {
	for _, allowed := range c.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
