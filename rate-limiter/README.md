# RateLimiter

A lightweight Go package for rate limiting using a token bucket approach.  
It allows a defined number of operations (`intervals`) over a given time span (`timeSpan`),  
and can be gracefully stopped using context or a manual call.

## Features

- ✅ Limit the number of actions over a defined duration  
- ✅ Context-aware cancellation  
- ✅ Simple API with minimal dependencies  
- ✅ Lightweight goroutine-based implementation  

## Installation

```bash
go get github.com/hiteshrepo/ratelimiter
```

## Usages

### using LimitCh(ctx)
```go
package main

import (
	"context"
	"fmt"
	"time"

	rlm "github.com/hiteshrepo/awesome-tools/rate-limiter"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// To rate limit 5 ops per second
	rl := rlm.NewRateLimiter(ctx, time.Second, 5)

	for i := 0; i < 10; i++ {
		select {
		case <-rl.LimitCh(ctx):
			fmt.Println("Do rate-limited work")
		default:
			fmt.Println("Rate limit exceeded")
			time.Sleep(200 * time.Millisecond)
		}
	}

	rl.Stop()
}
```

### using Wait(ctx) [RECOMMENDED]
```go
package main

import (
	"context"
	"fmt"
	"time"

	rlm "github.com/hiteshrepo/awesome-tools/rate-limiter"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// To rate limit 5 ops per second
	rl := rlm.NewRateLimiter(ctx, time.Second, 5)

	for i := 0; i < 10; i++ {
		if rl.Wait(ctx) {
			fmt.Println("Do rate-limited work")
		} else {
			fmt.Println("Context canceled or limiter stopped")
		}
	}

	rl.Stop()
}
```

## API

### NewRateLimiter(ctx context.Context, timeSpan time.Duration, intervals int) RateLimiter
Creates a rate limiter that allows intervals actions per timeSpan.
Internally uses a goroutine that adds and removes tokens at evenly spaced intervals.
Automatically stops when the provided ctx is canceled or Stop() is called.

### RateLimiter.LimitCh(ctx context.Context) <-chan struct{}
LimitCh returns the internal token channel used for rate limiting.
External callers can select on this channel to wait for token availability to proceed.

### RateLimiter.Wait(ctx context.Context) bool
Wait blocks until a token is available or the context is done.
Returns true if a token was acquired, false if context was canceled.
Prefer this over `LimitCh` method for safer and more idiomatic usage.

### RateLimiter.Stop()
Manually stops the rate limiter's internal goroutine.

