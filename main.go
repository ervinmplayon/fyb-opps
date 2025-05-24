package main

import (
	"time"

	"github.com/ervinmplayon/fyb-opps/ratelimiter"
)

var limiter = ratelimiter.NewLimiter(2 * time.Second)
