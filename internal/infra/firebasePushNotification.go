package infra

import (
	"context"
	"sms-gateway/internal/domain"

	"firebase.google.com/go/messaging"
)

type FirebasePushNotification struct {
	client *messaging.Client
	dryRun bool
}

func NewFirebasePushNotification(client *messaging.Client) FirebasePushNotification {
	return FirebasePushNotification{client: client, dryRun: false}
}

func (receiver *FirebasePushNotification) EnableDryRun() {
	receiver.dryRun = true
}

func (receiver *FirebasePushNotification) IsDryRun() bool {
	return receiver.dryRun
}

func (receiver *FirebasePushNotification) Send(ctx context.Context, message domain.Sms, token string) error {
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
		//_, err := receiver.client.SendDryRun(receiver.ctx, firebaseMessage)
		//return err
		return nil
	} else {
		_, err := receiver.client.Send(ctx, firebaseMessage)
		return err
	}
}
