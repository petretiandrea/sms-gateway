package domain

import (
	"context"
	"fmt"
)

type DeliveryNotificationConfig struct {
	AccountId  AccountID
	WebhookURL string
	Enabled    bool
}

type DeliveryNotificationConfigRepository interface {
	Save(context.Context, DeliveryNotificationConfig) (bool, error)
	FindById(context.Context, AccountID) *DeliveryNotificationConfig
}

func (config DeliveryNotificationConfig) String() string {
	return fmt.Sprintf(
		"Account Id %s, Enabled %t, WebhookURL %s",
		string(config.AccountId),
		config.Enabled,
		config.WebhookURL,
	)
}
