package scanner

import "time"

type RateLimiter struct {
	tokens chan struct{}
	ticker *time.Ticker
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	if requestsPerSecond <= 0 {
		return nil
	}

	rl := &RateLimiter{
		tokens: make(chan struct{}, requestsPerSecond),
		ticker: time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
	}

	for i := 0; i < requestsPerSecond; i++ {
		rl.tokens <- struct{}{}
	}

	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Wait() {
	if rl == nil {
		return
	}
	<-rl.tokens
}

func (rl *RateLimiter) Stop() {
	if rl == nil {
		return
	}
	rl.ticker.Stop()
}
