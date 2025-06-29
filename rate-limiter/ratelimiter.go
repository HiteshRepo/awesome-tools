package ratelimiter

import (
	"context"
	"time"
)

type RateLimiter struct {
	limitCh chan struct{}
	stopCh  chan struct{}
}

// NewRateLimiter creates a rate limiter allowing `rpi` actions per `interval`.
// It stops when the context is done or Stop() is called.
// timeSpan: The total time span in which rpi actions are allowed (e.g., 1 second, 1 minute, etc.).
// intervals: The number of intervals for the given timeSpan.
func NewRateLimiter(
	ctx context.Context,
	timeSpan time.Duration,
	intervals int) RateLimiter {
	limitCh := make(chan struct{}, intervals)
	stopCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(timeSpan / time.Duration(intervals))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				limitCh <- struct{}{}
			case <-ctx.Done():
				return
			case <-stopCh:
				return
			}
		}
	}()

	return RateLimiter{
		limitCh: limitCh,
		stopCh:  stopCh,
	}
}

// Wait blocks until a token is available or the context is done.
// Returns true if a token was acquired, false if context was canceled.
func (rl RateLimiter) Wait(ctx context.Context) bool {
	select {
	case <-rl.limitCh:
		return true
	case <-ctx.Done():
		return false
	}
}

// LimitCh returns the internal token channel used for rate limiting.
// External callers can select on this channel to wait for token availability to proceed.
// Prefer using the Wait method for safer and more idiomatic usage.
func (rl RateLimiter) LimitCh(ctx context.Context) <-chan struct{} {
	return rl.limitCh
}

// Stops the ratelimiter
func (rl RateLimiter) Stop() {
	rl.stopCh <- struct{}{}
}
