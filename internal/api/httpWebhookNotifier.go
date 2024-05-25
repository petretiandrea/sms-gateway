package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/generated/openapi"
)

type HttpWebhookNotifier struct {
}

func (h HttpWebhookNotifier) Notify(sms *domain.Sms, webhookUrl string) error {
	if notification, err := json.Marshal(openapi.SmsEntityResponse{
		Id:           string(sms.Id),
		To:           sms.To,
		From:         sms.From.Number,
		Content:      sms.Content,
		Owner:        string(sms.UserId),
		CreatedAt:    sms.CreatedAt,
		IsSent:       sms.IsSent,
		SendAttempts: int32(sms.SendAttempts),
	}); err != nil {
		return err
	} else {
		response, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(notification))
		if err != nil {
			return err
		}
		if response.StatusCode >= 300 {
			return fmt.Errorf("webhook endpoint response is not successfully, responseCode %d", response.StatusCode)
		}
		return nil
	}
}

var _ domain.WebhookNotifier = (*HttpWebhookNotifier)(nil)
