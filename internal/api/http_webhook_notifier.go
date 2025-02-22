package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sms-gateway/internal/domain"

	"github.com/google/uuid"
)

type HttpWebhookNotifier struct {
}

func (h HttpWebhookNotifier) Notify(ctx context.Context, sms *domain.Sms, webhookUrl string) error {
	eventType := mapNotificationEventType(*sms)
	if eventType == "" {
		return fmt.Errorf("cannot establish notification type")
	}
	notification := EventNotificationDto {
		EventType: eventType,
		Data: SmsEntityResponse{
			Id:          uuid.MustParse(string(sms.Id)),
			To:          sms.To,
			From:        sms.From.Number,
			Content:     sms.Content,
			Owner:       uuid.MustParse(string(sms.UserId)),
			CreatedAt:   sms.CreatedAt,
			IsSent:      sms.IsSent,
			LastAttempt: lastAttemptToDto(sms.LastAttempt),
			UpdatedAt:   sms.LastUpdateAt,
		},
		Metadata: sms.Metadata,
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

func mapNotificationEventType(sms domain.Sms) EventNotificationType {
	if sms.LastAttempt != nil {
		if _, ok := sms.LastAttempt.(domain.SuccessAttempt); ok {
			return MessageDeliverSucceeded
		} else if _, ok := sms.LastAttempt.(domain.FailedAttempt); ok {
			return MessageDeliverFailed
		}
	}
	return ""
}

var _ domain.WebhookNotifier = (*HttpWebhookNotifier)(nil)
