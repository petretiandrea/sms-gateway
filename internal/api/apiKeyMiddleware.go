package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
)

func NewApiKeyMiddleware(accountService application.UserAccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Api-Key")
		if apiKey != "" {
			if user := accountService.GetUserAccountByApiKey(domain.ApiKey(apiKey)); user != nil {
				c.Set("user", *user)
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, errors.New("invalid ApiKey"))
		return
	}
}
