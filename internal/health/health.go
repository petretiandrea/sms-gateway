package health

import (
	"context"
	"github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func RegisterGinHealthCheck(gin *gin.Engine, db Pinger) {
	gin.GET("/health", healthCheckHandler(createHealthCheck(db)))
}

func FilterHealthCheck(request *http.Request) bool {
	return !strings.Contains(request.URL.Path, "health")
}

func healthCheckHandler(healthCheck health.Checker) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		result := healthCheck.Check(ctx)
		if result.Status == health.StatusDown {
			ctx.JSONP(http.StatusInternalServerError, result)
		} else {
			ctx.JSONP(http.StatusOK, result)
		}
	}
}

func createHealthCheck(db Pinger) health.Checker {
	checker := health.NewChecker(
		health.WithCacheDuration(5*time.Minute),
		health.WithCheck(health.Check{
			Name:    "postgres",
			Timeout: 2 * time.Second,
			Check: func(ctx context.Context) error {
				return db.Ping(ctx)
			},
		}),
	)

	return checker
}
