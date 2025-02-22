package domain

import "context"

type WebhookNotifier interface {
	Notify(ctx context.Context, sms *Sms, webhookUrl string) error
}
