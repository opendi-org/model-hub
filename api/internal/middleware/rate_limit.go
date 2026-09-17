package middleware

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	"github.com/gin-gonic/gin"
)

// NewLimiter creates a new rate limiter for an endpoint or set of endpoints, using expected conventions for this
// project (separate Burst and Max values, expiration TTL of 1).
//   - burst is the maximum number of requests allowed through in a short time.
//   - requestsPerMinute is the rate at which requests are allowed over a sustained period.
func NewLimiter(burst int, requestsPerMinute float64) *limiter.Limiter {
	return tollbooth.NewLimiter(requestsPerMinute/60, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour}).SetBurst(burst)
}

// CheckRateLimit uses a Limiter to enforce a rate limit keyed to client IP and the requested endpoint path.
//
// If the client has made too many requests to the requested endpoint, sets "Retry-After" header with the
// number of seconds to wait for the request to succeed.
func CheckRateLimit(limit *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if httpErr := tollbooth.LimitByKeys(limit, []string{c.ClientIP(), c.FullPath()}); httpErr != nil {
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(1/limit.GetMax()))))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded. Please wait before retrying."})
			c.Abort()
			return
		}

		c.Next()
	}
}
