package postgres

import (
	"context"
	"sms-gateway/internal/domain"
)

type DeliveryNotificationRepository struct {
	db DBTX
}

func NewDeliveryNotificationRepository(db DBTX) DeliveryNotificationRepository {
	return DeliveryNotificationRepository{db: db}
}

func (r DeliveryNotificationRepository) Save(ctx context.Context, config domain.DeliveryNotificationConfig) (bool, error) {
	_, err := r.db.Exec(ctx, `
INSERT INTO delivery_notification_configs (account_id, webhook_url, enabled)
VALUES ($1, $2, $3)
ON CONFLICT (account_id) DO UPDATE SET
    webhook_url = EXCLUDED.webhook_url,
    enabled = EXCLUDED.enabled`,
		string(config.AccountId),
		config.WebhookURL,
		config.Enabled,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r DeliveryNotificationRepository) FindById(ctx context.Context, id domain.AccountID) *domain.DeliveryNotificationConfig {
	row := r.db.QueryRow(ctx, `
SELECT account_id, webhook_url, enabled
FROM delivery_notification_configs
WHERE account_id = $1`, string(id))

	var config domain.DeliveryNotificationConfig
	var accountID string
	if err := row.Scan(&accountID, &config.WebhookURL, &config.Enabled); err != nil {
		return nil
	}

	config.AccountId = domain.AccountID(accountID)
	return &config
}

var _ domain.DeliveryNotificationConfigRepository = (*DeliveryNotificationRepository)(nil)
