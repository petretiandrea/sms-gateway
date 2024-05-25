package health

import (
	"cloud.google.com/go/firestore"
	"github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

func RegisterGinHealthCheck(gin *gin.Engine, client *firestore.Client) {
	gin.GET("/health", healthCheckHandler(createHealthCheck(client)))
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

func createHealthCheck(client *firestore.Client) health.Checker {
	firebaseHealth := NewFirebaseHealth(client)
	checker := health.NewChecker(
		health.WithCacheDuration(5*time.Minute),
		health.WithCheck(health.Check{
			Name:    "firebase",
			Timeout: 2 * time.Second,
			Check:   firebaseHealth.Check,
		}),
	)

	return checker
}
