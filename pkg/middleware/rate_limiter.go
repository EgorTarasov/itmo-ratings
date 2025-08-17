package middleware

import (
	"net/http"
	"time"
)

type RateLimiter struct {
	next   http.Handler
	tokens chan struct{}
}

func NewRateLimiter(next http.Handler, rps, burst int) http.Handler {
	if rps <= 0 {
		rps = 1
	}
	if burst < 1 {
		burst = 1
	}
	rl := &RateLimiter{
		next:   next,
		tokens: make(chan struct{}, burst),
	}

	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(rps)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// bucket full; drop token
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case <-rl.tokens:
		rl.next.ServeHTTP(w, r)
	default:
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("too many requests"))
	}
}
