package middleware

import (
	"errors"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"

	"sms-gateway/internal/api"

	"github.com/gin-gonic/gin"
)


func NewApiKeyMiddleware(accountService application.UserAccountService) api.StrictMiddlewareFunc {
	return func (f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx *gin.Context, request interface{}) (response interface{}, err error) {
			if _, exists := ctx.Get(api.ApiKeyAuthScopes); exists {
				apiKey := ctx.GetHeader("Api-Key")
				authorized := false
				if apiKey != "" {
					if user := accountService.GetUserAccountByApiKey(ctx, domain.ApiKey(apiKey)); user != nil {
						ctx.Set("user", *user)
						authorized = true
					}
				}
				if !authorized {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, errors.New("invalid ApiKey"))
					return
				}
			}
			return f(ctx, request)
		}
	}
}