package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
)

type Filter func(*http.Request) bool

func anyMatch(request *http.Request, f []Filter) bool {
	for _, filter := range f {
		if filter(request) {
			return true
		}
	}
	return false
}

func NewApiKeyMiddleware(accountService application.UserAccountService, f ...Filter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !anyMatch(c.Request, f) {
			c.Next()
			return
		}

		apiKey := c.GetHeader("Api-Key")
		if apiKey != "" {
			if user := accountService.GetUserAccountByApiKey(domain.ApiKey(apiKey)); user != nil {
				c.Set("user", *user)
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, errors.New("invalid ApiKey"))
		return
	}
}
