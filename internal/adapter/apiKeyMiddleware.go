package adapter

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/account"
)

func NewApiKeyMiddleware(accountService account.UserAccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.Param("Api-Key")
		if apiKey != "" {
			if user := accountService.GetUserAccountByApiKey(account.ApiKey(apiKey)); user != nil {
				c.Set("user", user)
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, errors.New("invalid ApiKey"))
		return
	}
}
