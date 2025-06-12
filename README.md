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

## Expiring Old IPs
* The goal is to avoid memory bloat by cleaning up IPs that haven't sent requests in a while
* Track `lastSeen` for each limiter
* Run a background goroutine that periodically removes stale entries
* Each IP gets its own limiter
* Each limiter tracks `lastSeen`
* A goroutine periodically checks for stale IPs and removes them
* This implementation is fully thread safe

## Iterating (range) over ticker.C deep dive
The `time.Ticker` has a field `C` which is a channel that sends a value at each tick interval. The `range` keyword can iterate over channels, receiving values sent on the channel until the channel is closed.
When I use `for range ticker.C`, the loop will block and wait for a value to be sent on the `ticker.C` channel. After receiving a value, the loop will execute its body, and then it will wait again for the next value. This makes `range` a convenient way to continuously execute code at regular time intervals defined by the ticker.
The loop terminates when the ticker is stopped using `ticker.Stop()` which closes the channel. 

## High Concurrency
High concurrency enables many clients (or goroutines) to send requests at the same time. This is common in:
* Public APIs
* Internal services with traffic spikes
* Systems behind load balancers or gateways
* Anything deployed on ECS, K8s, etc
### What does this introduce?
#### 1. Contention on Shared Resources
Current implementation uses `mu sync.Mutex`. When thousands of requests hit simultaneously, they all try to lock `mu` in `getlimiter(ip)`.
```
lm.mu.Lock()
defer lm.mu.Unlock()
```
This will result in lock contention, which slows everyone down -- they sit around waiting to acquire the mutex.
#### 2.Risk of Latency Spikes
In real-time apps (APIs, ad bidding, etc) a delay of even tens of milliseconds matters. Mutex contention in high concurrency environments can:
* Introduce inconsistent response times
* Block unrelated requests just because they access the same `LimiterMap`.

#### 3. Scalability Benefits of Lock-Free Access
Using `sync.Map` avoids traditional locks for reads and writes by using atomic operations and internal sharding under the hood. That means:
* Reads are lock-free
* Writes are optimized (especially after `LoadOrStore` warm-up)
* Multiple goroutines can access/update independently without blocking each other. 
### Summary
With Mutex:
* Slower under load
* Goroutines block each other
* Simpler logic
* Suitable for low traffic

With `sync.Map`
* Optimized for high concurrency
* Goroutines run independently
* Slightly more complex logic
* Preferred at scale. 

## Future Feature Implementations
- [x] Immediately hard throttle and reject requests if they arrive too fast, instead of waiting. This requires a change in design from a token bucket that just throttles to a leaky bucket or fixed window that enforces a limit per interval with no waiting. 
- [x] Per-client limiting: Use a `map[string]*HardThrottleLimiter` keyed by IP or token
- [x] Add `cleaner` goroutine that expires old IPs
- [x] Optimize for high concurrency (e.g., using `sync.Map`)
- [ ] `Allow` method to return the remaining quota 
- [ ] Logging: Add logging for every rejection if needed
- [ ] Ticket: integrate logger into hard throttle limiter
- [ ] Ticket: integrate logger into limiter map
- [ ] Logging added to the cleanup process. 
- [ ] Sliding window: More precise throttling (less bursty)
- [ ] Add X-Forwarded-For support (for use behind reverse proxies) 

## Concepts to Learn
- [ ] Understand how this implementation is different from a `gin` router setup implementation 
- [ ] Double-Checked Locking: A Concurrency-Safe Pattern



