package postgres

import (
	"context"
	"encoding/json"
	"sms-gateway/internal/domain"
	"time"
)

type DLQMessageRepository struct {
	db DBTX
}

func NewDLQMessageRepository(db DBTX) DLQMessageRepository {
	return DLQMessageRepository{db: db}
}

func (r DLQMessageRepository) Save(ctx context.Context, message domain.DLQMessage) error {
	metadataMap := message.Metadata
	if metadataMap == nil {
		metadataMap = map[string]string{}
	}

	metadata, err := json.Marshal(metadataMap)
	if err != nil {
		return err
	}

	var occurredAt any
	if !message.OccurredAt.IsZero() {
		occurredAt = message.OccurredAt
	}

	createdAt := message.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err = r.db.Exec(ctx, `
INSERT INTO dlq_messages (
    message_id, channel, exchange, routing_key, content_type,
    payload, metadata, occurred_at, created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9
)`,
		message.MessageID,
		message.Channel,
		message.Exchange,
		message.RoutingKey,
		message.ContentType,
		message.Payload,
		metadata,
		occurredAt,
		createdAt,
	)
	return err
}

var _ domain.DLQMessageRepository = (*DLQMessageRepository)(nil)
