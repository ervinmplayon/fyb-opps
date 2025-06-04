# fyb-opps
Rate Limiter using Context

## What does `Wait(ctx)` actually do?
* The limiter never explicitly rejects requests. It just makes them wait for the next available tick.
* Only `ctx.Done()` (case 2) or `l.quit` (case 3) cause `Wait()` to return an error.
    - IF the client disconnects, the context is cancelled (case 2).
    - IF I set a timeout(`context.WithTimeout`) and the wait exceeds it, the context expires (case 2).
    - IF the app manually shuts down and closes `l.quit`, it returns early (case 3).

## When is "too many requests" triggered?
It is only triggered when the context times out before the ticker has a chance to release a token. Thats when `ctx.Done()` fires before `<-l.ticker.C` becomes available - and the handler returns a 429. 
* If 1 request is allowed every 2 seconds...
* But the client makes 5 requests in 2 seconds...
* And `limiter.Wait(cx)` is wrapped in `context.WithTimeout(ctx, 1*time.Second)`...
* Then some of those requests will timeout before a token is ready and will return a 429.

## Hard Throttle Feature
* Requests within 5 per second succeed.
* Subsequent requests within the same second gets a `429 Too Many Requests`
* On the next second, the count resets.
* Listens on more complex endpoints

## Per IP Address Limiter Version
* Tracks each client by IP address
* Allow each client up to `N` requests per interval
* Reject clients who exceed the limit immediately (hard throttle)

## Future Feature Implementations
- [x] Immediately hard throttle and reject requests if they arrive too fast, instead of waiting. This requires a change in design from a token bucket that just throttles to a leaky bucket or fixed window that enforces a limit per interval with no waiting. 
- [x] Per-client limiting: Use a `map[string]*HardThrottleLimiter` keyed by IP or token
- [ ] Sliding window: More precise throttling (less bursty)
- [ ] Logging: Add logging for every rejection if needed
- [ ] Ticket: integrate logger into hard throttle limiter
- [ ] Ticket: integrate logger into limiter map
- [ ] Add goroutine that expires old IPs
- [ ] Add X-Forwarded-For support (for use behind reverse proxies) 
