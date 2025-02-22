package api

import (
	"context"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
)

type DeliveryNotificationController struct {
	Account              application.UserAccountService
	DeliveryNotification application.DeliveryNotificationService
}

// PostWebhooks implements StrictServerInterface.
func (c *DeliveryNotificationController) PostWebhooks(
	ctx context.Context, 
	request PostWebhooksRequestObject,
) (PostWebhooksResponseObject, error) {
	user := ctx.Value("user").(domain.UserAccount)
	if request.Body.Enabled && request.Body.DefaultWebhookUrl == nil {
		return PostWebhooks200JSONResponse{}, nil
	}
	configToUpdate := domain.DeliveryNotificationConfig{
		WebhookURL: *request.Body.DefaultWebhookUrl,
		Enabled:    request.Body.Enabled,
		AccountId:  user.Id,
	}
	if config, err := c.DeliveryNotification.UpdateDeliveryConfig(ctx, configToUpdate); err == nil {
		return PostWebhooks200JSONResponse{
			WebhookURL: &config.WebhookURL,
			Enabled: config.Enabled,
		}, nil
	} else {
		return PostWebhooks400Response{}, nil
	}
}

