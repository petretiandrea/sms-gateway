package postgres

import (
	"context"
	"fmt"
	"sms-gateway/internal/domain"
	"strings"

	"github.com/jackc/pgx/v5"
)

type MessageRepository struct {
	db DBTX
}

func NewMessageRepository(db DBTX) MessageRepository {
	return MessageRepository{db: db}
}

func (r MessageRepository) Find(ctx context.Context, params domain.QueryParams) ([]domain.Sms, error) {
	query := strings.Builder{}
	query.WriteString(`
SELECT id, from_phone, to_phone, content, is_sent,
       last_attempt_type, last_attempt_phone_id, last_attempt_count, last_attempt_failure_reason,
       owner_id, idempotency_key, created_at, updated_at, metadata, webhook_url
FROM sms_messages`)

	var args []any
	var filters []string
	if params.From != "" {
		args = append(args, params.From)
		filters = append(filters, fmt.Sprintf("from_phone = $%d", len(args)))
	}
	if params.IsSent != nil {
		args = append(args, *params.IsSent)
		filters = append(filters, fmt.Sprintf("is_sent = $%d", len(args)))
	}
	if len(filters) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(filters, " AND "))
	}

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Sms
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r MessageRepository) Save(ctx context.Context, message domain.Sms) (*domain.Sms, error) {
	entity, err := messageToEntity(message)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(ctx, `
INSERT INTO sms_messages (
    id, from_phone, to_phone, content, is_sent,
    last_attempt_type, last_attempt_phone_id, last_attempt_count, last_attempt_failure_reason,
    owner_id, idempotency_key, created_at, updated_at, metadata, webhook_url
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15
)
ON CONFLICT (id) DO UPDATE SET
    from_phone = EXCLUDED.from_phone,
    to_phone = EXCLUDED.to_phone,
    content = EXCLUDED.content,
    is_sent = EXCLUDED.is_sent,
    last_attempt_type = EXCLUDED.last_attempt_type,
    last_attempt_phone_id = EXCLUDED.last_attempt_phone_id,
    last_attempt_count = EXCLUDED.last_attempt_count,
    last_attempt_failure_reason = EXCLUDED.last_attempt_failure_reason,
    owner_id = EXCLUDED.owner_id,
    idempotency_key = EXCLUDED.idempotency_key,
    updated_at = EXCLUDED.updated_at,
    metadata = EXCLUDED.metadata,
    webhook_url = EXCLUDED.webhook_url`,
		entity.id,
		entity.from,
		entity.to,
		entity.content,
		entity.isSent,
		entity.lastAttemptType,
		entity.lastAttemptPhoneID,
		entity.lastAttemptCount,
		entity.lastAttemptFailureReason,
		entity.owner,
		entity.idempotencyKey,
		entity.createdAt,
		entity.updatedAt,
		entity.metadata,
		entity.webhookURL,
	)
	if err != nil {
		return nil, err
	}

	message.LastUpdateAt = entity.updatedAt
	return &message, nil
}

func (r MessageRepository) FindById(ctx context.Context, id domain.SmsId) *domain.Sms {
	message, err := scanMessage(r.db.QueryRow(ctx, `
SELECT id, from_phone, to_phone, content, is_sent,
       last_attempt_type, last_attempt_phone_id, last_attempt_count, last_attempt_failure_reason,
       owner_id, idempotency_key, created_at, updated_at, metadata, webhook_url
FROM sms_messages
WHERE id = $1`, string(id)))
	if err != nil {
		return nil
	}
	return message
}

func (r MessageRepository) FindExisting(ctx context.Context, idempotencyKey string) *domain.Sms {
	message, err := scanMessage(r.db.QueryRow(ctx, `
SELECT id, from_phone, to_phone, content, is_sent,
       last_attempt_type, last_attempt_phone_id, last_attempt_count, last_attempt_failure_reason,
       owner_id, idempotency_key, created_at, updated_at, metadata, webhook_url
FROM sms_messages
WHERE idempotency_key = $1`, idempotencyKey))
	if err != nil {
		return nil
	}
	return message
}

func scanMessage(row pgx.Row) (*domain.Sms, error) {
	var entity messageEntity
	err := row.Scan(
		&entity.id,
		&entity.from,
		&entity.to,
		&entity.content,
		&entity.isSent,
		&entity.lastAttemptType,
		&entity.lastAttemptPhoneID,
		&entity.lastAttemptCount,
		&entity.lastAttemptFailureReason,
		&entity.owner,
		&entity.idempotencyKey,
		&entity.createdAt,
		&entity.updatedAt,
		&entity.metadata,
		&entity.webhookURL,
	)
	if err != nil {
		return nil, err
	}
	return entity.toMessage()
}

var _ domain.Repository = (*MessageRepository)(nil)
