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

	/*
	 * This HandleFunc syntax is a Go idiom, I am passing the function handler as a value, not calling it.
	 * http.HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	 * expects a function reference that matches the func(http.ResponseWriter, *http.Request) signature.
	 * So passing the handler without () is passing the function definition itself, not executing it.
	 * Go functions are 1st-class values, they can be assigned to variables or passed as arguments like any other data.
	 */
	http.HandleFunc("/", handler)
	log.Println("Serving on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
