package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/generated/openapi"
)

type DeliveryNotificationController struct {
	Account              application.UserAccountService
	DeliveryNotification application.DeliveryNotificationService
}

func (d DeliveryNotificationController) WebhooksPost(c *gin.Context) {
	var request openapi.CreateWebhookRequest
	user := c.MustGet("user").(domain.UserAccount)
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	configToUpdate := domain.DeliveryNotificationConfig{
		WebhookURL: request.DefaultWebhookUrl,
		Enabled:    request.Enabled,
		AccountId:  user.Id,
	}
	if config, err := d.DeliveryNotification.UpdateDeliveryConfig(configToUpdate); err == nil {
		c.JSONP(http.StatusOK, openapi.WebhookEntityResponse{
			WebhookURL: config.WebhookURL,
			Enabled:    config.Enabled,
		})
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

var (
	_ openapi.WebhooksAPI = (*DeliveryNotificationController)(nil)
)
