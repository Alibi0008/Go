package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"practice7/internal/utils"

	"github.com/gin-gonic/gin"
)

type visitorWindow struct {
	Count   int
	ResetAt time.Time
}

func NewRateLimiter(jwtManager *utils.JWTManager, maxRequests int, window time.Duration) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitorWindow)
	)

	return func(c *gin.Context) {
		key := resolveRateLimitKey(c, jwtManager)
		now := time.Now()

		mu.Lock()
		record, exists := visitors[key]
		if !exists || now.After(record.ResetAt) {
			record = &visitorWindow{
				Count:   0,
				ResetAt: now.Add(window),
			}
			visitors[key] = record
		}

		record.Count++
		remaining := maxRequests - record.Count
		resetInSeconds := int(time.Until(record.ResetAt).Seconds())
		if resetInSeconds < 0 {
			resetInSeconds = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(remaining, 0)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetInSeconds))

		if record.Count > maxRequests {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		mu.Unlock()

		c.Next()
	}
}

func resolveRateLimitKey(c *gin.Context, jwtManager *utils.JWTManager) string {
	tokenString := extractBearerToken(c.GetHeader("Authorization"))
	if tokenString != "" {
		claims, err := jwtManager.ParseToken(tokenString)
		if err == nil && claims.UserID != "" {
			return "user:" + claims.UserID
		}
	}

	return "ip:" + c.ClientIP()
}
