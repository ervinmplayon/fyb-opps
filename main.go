package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ervinmplayon/fyb-opps/ratelimiter"
)

var limiter = ratelimiter.NewLimiter(2 * time.Second)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := limiter.Wait(ctx); err != nil {
		http.Error(w, "Rate limit exceeded or context canceled", http.StatusTooManyRequests)
		return
	}

	fmt.Fprintln(w, "Request allowed at", time.Now())
}

func main() {
	defer limiter.Stop() // * Clean up the ticker

	http.HandleFunc("/", handler) // * this a Go idiom - you are passing the function handler as a value, not calling it.
	log.Println("Serving on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
