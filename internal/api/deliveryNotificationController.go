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

func (d DeliveryNotificationController) WebhookDisablePut(c *gin.Context) {
	user := c.MustGet("user").(domain.UserAccount)
	if config := d.DeliveryNotification.DisableDeliveryNotification(
		user.Id,
	); config == nil {
		c.JSON(http.StatusNotFound, "Configuration not found")
	} else {
		c.JSONP(http.StatusOK, openapi.WebhookEntityResponse{
			WebhookURL: config.WebhookURL,
			Enabled:    config.Enabled,
		})
	}
}

func (d DeliveryNotificationController) WebhookEnablePut(c *gin.Context) {
	var request openapi.EnableWebhookRequest
	user := c.MustGet("user").(domain.UserAccount)
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	configToUpdate := domain.DeliveryNotificationConfig{
		WebhookURL: request.WebhookURL,
		Enabled:    true,
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
