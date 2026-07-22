package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple in-memory rate limiter using a token bucket approach.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     int
	burst    int
}

type visitor struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a rate limiter that allows `rate` requests per second with burst capacity.
func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
	}
	// Clean up old entries periodically
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastCheck) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:    rl.burst - 1,
			lastCheck: time.Now(),
		}
		return true
	}

	elapsed := time.Since(v.lastCheck)
	v.tokens += int(elapsed.Seconds()) * rl.rate
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}
	v.lastCheck = time.Now()

	if v.tokens > 0 {
		v.tokens--
		return true
	}
	return false
}

// Limit returns middleware that rate-limits requests.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !rl.allow(ip) {
			http.Error(w, `{"success":false,"message":"Terlalu banyak permintaan. Coba lagi nanti."}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
