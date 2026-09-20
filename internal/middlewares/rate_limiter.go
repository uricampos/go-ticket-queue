package middlewares

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
	IPList  map[string]*rate.Limiter
	IPMutex *sync.Mutex
}

func NewRateLimitMiddleware(ipList map[string]*rate.Limiter, ipMutex *sync.Mutex) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		IPList:  ipList,
		IPMutex: ipMutex,
	}
}

func (rlm *RateLimitMiddleware) getVisitor(ip string) *rate.Limiter {
	rlm.IPMutex.Lock()
	defer rlm.IPMutex.Unlock()

	limiter, exists := rlm.IPList[ip]

	if !exists {
		limiter = rate.NewLimiter(2, 15)
		rlm.IPList[ip] = limiter
	}

	return limiter

}

func (rlm *RateLimitMiddleware) GetRateLimit() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		clientIP := ctx.ClientIP()

		limiter := rlm.getVisitor(clientIP)

		if limiter.Allow() == false {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
