package domain

type WebhookNotifier interface {
	Notify(sms *Sms, webhookUrl string) error
}
