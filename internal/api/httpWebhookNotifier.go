package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sms-gateway/internal/domain"
)

type HttpWebhookNotifier struct {
}

func (h HttpWebhookNotifier) Notify(sms *domain.Sms, webhookUrl string) error {
	if notification, err := json.Marshal(SmsEntityResponse{
		Id:           string(sms.Id),
		To:           sms.To,
		From:         sms.From.Number,
		Content:      sms.Content,
		UserId:       string(sms.UserId),
		CreatedAt:    sms.CreatedAt,
		IsSent:       sms.IsSent,
		SendAttempts: sms.SendAttempts,
	}); err != nil {
		return err
	} else {
		response, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(notification))
		if err != nil {
			return err
		}
		if response.StatusCode != 200 {
			return errors.New("webhook endpoint response is not successfully")
		}
		return nil
	}
}

var _ domain.WebhookNotifier = (*HttpWebhookNotifier)(nil)
