package monitor

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientWindow struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientWindow
	limit   int
	window  time.Duration
}

// in-memory、per-client、fixed-window rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*clientWindow),
		limit:   limit,
		window:  window,
	}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// skip Kubernetes health probes
		if r.URL.Path == "/api/v1/health" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		clientID := clientIP(r)

		now := time.Now()

		l.mu.Lock()

		entry, exists := l.clients[clientID]

		if !exists || now.Sub(entry.windowStart) >= l.window {
			entry = &clientWindow{
				count:       0,
				windowStart: now,
			}

			l.clients[clientID] = entry
		}

		entry.count++

		exceeded := entry.count > l.limit

		l.mu.Unlock()

		if exceeded {
			RateLimitExceededTotal.Inc()

			http.Error(
				w,
				"rate limit exceeded",
				http.StatusTooManyRequests,
			)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
