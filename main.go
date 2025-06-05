package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/ervinmplayon/fyb-opps/ratelimiter"
)

var defaultLimiter = ratelimiter.NewLimiter(2 * time.Second)
var hardThrottleLimiter = ratelimiter.NewHardThrottleLimiter(5, time.Second)
var ipLimiter = ratelimiter.NewLimiterMap(5, time.Second)

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // * fallback
	}
	return ip
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := defaultLimiter.Wait(ctx); err != nil {
		http.Error(w, "Rate limit exceeded or context canceled", http.StatusTooManyRequests)
		return
	}

	fmt.Fprintln(w, "Request allowed at", time.Now())
}

func hardThrottleHandler(w http.ResponseWriter, r *http.Request) {
	if !hardThrottleLimiter.Allow() {
		http.Error(w, "Rate limit Exceeded", http.StatusTooManyRequests)
		return
	}

	fmt.Fprintln(w, "OK - You passed the vibe check")
}

func ipLimiterHandler(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	if !ipLimiter.Allow(ip) {
		http.Error(w, "Rate Limit exceeded", http.StatusTooManyRequests)
		return
	}

	fmt.Fprintf(w, "Msg from %s. You're within the rate limit.\n", ip)
}

func main() {
	// * Clean up the ticker
	defer defaultLimiter.Stop()

	// * goroutine cleanup: remove IPs not seen in 5 mins -- goroutine frequency: every 1 min
	ipLimiter.StartCleanup(5*time.Minute, 1*time.Minute)

	/*
	 * This HandleFunc syntax is a Go idiom, I am passing the function handler as a value, not calling it.
	 * http.HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	 * expects a function reference that matches the func(http.ResponseWriter, *http.Request) signature.
	 * So passing the handler without () is passing the function definition itself, not executing it.
	 * Go functions are 1st-class values, they can be assigned to variables or passed as arguments like any other data.
	 */
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/not-a-honeypot", hardThrottleHandler)
	http.HandleFunc("/api", hardThrottleHandler)
	http.HandleFunc("/exclusive-content", ipLimiterHandler)
	log.Println("Serving on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
