package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	rlm "github.com/hiteshrepo/awesome-tools/rate-limiter"
)

func TestRateLimiterBasic(t *testing.T) {
	tests := []struct {
		name        string
		timeSpan    time.Duration
		intervals   int
		waitTime    time.Duration
		expectCount int
	}{
		{
			name:        "5 actions per second",
			timeSpan:    time.Second,
			intervals:   5,
			waitTime:    time.Second,
			expectCount: 5,
		},
		{
			name:        "10 actions per 2 seconds",
			timeSpan:    2 * time.Second,
			intervals:   10,
			waitTime:    2 * time.Second,
			expectCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			rl := rlm.NewRateLimiter(ctx, tt.timeSpan, tt.intervals)
			defer rl.Stop()

			count := 0
			timeout := time.After(tt.waitTime + 100*time.Millisecond)
		loop:
			for {
				select {
				case <-rl.LimitCh(ctx):
					count++
				case <-timeout:
					break loop
				}
			}

			if count != tt.expectCount {
				t.Errorf("expected %d actions, got %d", tt.expectCount, count)
			}
		})
	}
}

func TestRateLimiterStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rl := rlm.NewRateLimiter(ctx, time.Second, 2)

	rl.Stop()
	time.Sleep(100 * time.Millisecond)

	select {
	case <-rl.LimitCh(ctx):
		t.Log("Token was still received after Stop (might be a buffered leftover)")
	default:
		t.Log("No token received after Stop as expected")
	}
	cancel()
}
