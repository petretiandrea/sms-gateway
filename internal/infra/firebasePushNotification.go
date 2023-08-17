package infra

import (
	"context"
	"firebase.google.com/go/messaging"
	"sms-gateway/internal/domain"
)

type FirebasePushNotification struct {
	ctx    context.Context
	client *messaging.Client
	dryRun bool
}

func NewFirebasePushNotification(ctx context.Context, client *messaging.Client) FirebasePushNotification {
	return FirebasePushNotification{ctx: ctx, client: client, dryRun: false}
}

func (receiver *FirebasePushNotification) EnableDryRun() {
	receiver.dryRun = true
}

func (receiver *FirebasePushNotification) Send(message domain.Sms, token string) error {
	firebaseMessage := &messaging.Message{
		Token: token,
		Data: map[string]string{
			"id":      string(message.Id),
			"content": message.Content,
			"to":      message.To,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}

	if receiver.dryRun {
		_, err := receiver.client.Send(receiver.ctx, firebaseMessage)
		return err
	} else {
		_, err := receiver.client.SendDryRun(receiver.ctx, firebaseMessage)
		return err
	}
}
