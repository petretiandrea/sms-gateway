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
	smsEntityResponse := openapi.SmsEntityResponse{
		Id:        string(sms.Id),
		To:        sms.To,
		From:      sms.From.Number,
		Content:   sms.Content,
		Owner:     string(sms.UserId),
		CreatedAt: sms.CreatedAt,
		IsSent:    sms.IsSent,
	}
	if lastAttempt := lastAttemptToDto(sms.LastAttempt); lastAttempt != nil {
		smsEntityResponse.LastAttempt = *lastAttempt
	}
	notification := openapi.EventNotificationDto{
		EventType: openapi.MESSAGE_DELIVERED,
		Data:      smsEntityResponse,
		Metadata:  sms.Metadata,
	}
	if notification, err := json.Marshal(notification); err != nil {
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
