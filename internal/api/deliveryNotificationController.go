package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
)

type EnableDeliveryNotificationRequest struct {
	WebhookURL string `json:"webhookURL" binding:"required"`
}

type DeliveryNotificationConfigResponse struct {
	WebhookURL string `json:"webhookURL" binding:"required"`
	Enabled    bool   `json:"enabled" binding:"required"`
}

type DeliveryNotificationController struct {
	Account              application.UserAccountService
	DeliveryNotification application.DeliveryNotificationService
}

func (controller *DeliveryNotificationController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/webhook")
	router.Use(NewApiKeyMiddleware(controller.Account))
	router.PUT("/enable", controller.enableWebhookNotification)
	router.PUT("/disable", controller.disableWebhookNotification)
}

func (controller *DeliveryNotificationController) enableWebhookNotification(context *gin.Context) {
	var request EnableDeliveryNotificationRequest
	user := context.MustGet("user").(domain.UserAccount)
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	configToUpdate := domain.DeliveryNotificationConfig{
		WebhookURL: request.WebhookURL,
		Enabled:    true,
		AccountId:  user.Id,
	}
	if config, err := controller.DeliveryNotification.UpdateDeliveryConfig(configToUpdate); err == nil {
		context.JSONP(http.StatusOK, DeliveryNotificationConfigResponse{
			WebhookURL: config.WebhookURL,
			Enabled:    config.Enabled,
		})
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
}

func (controller *DeliveryNotificationController) disableWebhookNotification(context *gin.Context) {
	user := context.MustGet("user").(domain.UserAccount)
	if config := controller.DeliveryNotification.DisableDeliveryNotification(
		user.Id,
	); config == nil {
		context.JSON(http.StatusNotFound, "Configuration not found")
	} else {
		context.JSONP(http.StatusOK, DeliveryNotificationConfigResponse{
			WebhookURL: config.WebhookURL,
			Enabled:    config.Enabled,
		})
	}
}
