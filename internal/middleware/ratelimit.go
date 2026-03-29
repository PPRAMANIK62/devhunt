package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	max      int
	window   time.Duration
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}

	cleanupInterval := min(window, time.Minute)

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
			cutoff := time.Now().Add(-rl.window)
			for ip, reqs := range rl.requests {
				var valid []time.Time
				for _, t := range reqs {
					if t.After(cutoff) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.requests, ip)
				} else {
					rl.requests[ip] = valid
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window)
	var valid []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.requests[ip] = valid

	if len(rl.requests[ip]) >= rl.max {
		return false
	}
	rl.requests[ip] = append(rl.requests[ip], time.Now())
	return true
}

// RateLimit returns middleware that limits to `max` requests per `window` per IP.
// trustedProxies is an optional list of CIDR ranges (e.g. "10.0.0.0/8", "127.0.0.1/32")
// whose X-Forwarded-For header is trusted for real-client IP extraction.
// When deployed behind a reverse proxy, pass its CIDR here; otherwise leave empty
// to always use the direct connection IP (safe default that prevents header spoofing).
func RateLimit(max int, window time.Duration, trustedProxies ...string) func(http.Handler) http.Handler {
	rl := NewRateLimiter(max, window)

	var trustedNets []*net.IPNet
	for _, cidr := range trustedProxies {
		// Allow bare IPs without a mask.
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("ratelimit: invalid trusted proxy CIDR: " + cidr)
		}
		trustedNets = append(trustedNets, ipNet)
	}

	isTrusted := func(ip net.IP) bool {
		for _, n := range trustedNets {
			if n.Contains(ip) {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always resolve the direct connection IP first.
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			ip := host

			// Only trust X-Forwarded-For when the direct connection is a known proxy.
			// Take the rightmost entry — that's the IP the trusted proxy observed,
			// not the leftmost which the client can freely forge.
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" && isTrusted(net.ParseIP(host)) {
				parts := strings.Split(fwd, ",")
				ip = strings.TrimSpace(parts[len(parts)-1])
			}

			if !rl.Allow(ip) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "rate limit exceeded",
					"code":  "RATE_LIMIT",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
