package postgres

import (
	"encoding/json"
	"sms-gateway/internal/domain"
	"time"
)

type messageEntity struct {
	id                       string
	from                     string
	to                       string
	content                  string
	isSent                   bool
	lastAttemptType          *string
	lastAttemptPhoneID       *string
	lastAttemptCount         *int32
	lastAttemptFailureReason *string
	owner                    string
	idempotencyKey           string
	createdAt                time.Time
	updatedAt                time.Time
	metadata                 []byte
	webhookURL               *string
}

func messageToEntity(message domain.Sms) (messageEntity, error) {
	metadata, err := json.Marshal(message.Metadata)
	if err != nil {
		return messageEntity{}, err
	}

	updatedAt := time.Now()
	if !message.LastUpdateAt.IsZero() && message.LastAttempt == nil {
		updatedAt = message.LastUpdateAt
	}

	entity := messageEntity{
		id:             string(message.Id),
		from:           message.From.Number,
		to:             message.To,
		content:        message.Content,
		isSent:         message.IsSent,
		owner:          string(message.UserId),
		idempotencyKey: message.IdempotencyKey,
		createdAt:      message.CreatedAt,
		updatedAt:      updatedAt,
		metadata:       metadata,
		webhookURL:     message.Configuration.Url,
	}

	if message.CreatedAt.IsZero() {
		entity.createdAt = time.Now()
	}

	applyAttemptToEntity(&entity, message.LastAttempt)

	return entity, nil
}

func (entity messageEntity) toMessage() (*domain.Sms, error) {
	var metadata map[string]string
	if len(entity.metadata) > 0 {
		if err := json.Unmarshal(entity.metadata, &metadata); err != nil {
			return nil, err
		}
	}
	if metadata == nil {
		metadata = make(map[string]string)
	}

	return &domain.Sms{
		Id:             domain.SmsId(entity.id),
		From:           domain.PhoneNumber{Number: entity.from},
		To:             entity.to,
		Content:        entity.content,
		UserId:         domain.AccountID(entity.owner),
		IsSent:         entity.isSent,
		LastAttempt:    entityToAttempt(entity),
		CreatedAt:      entity.createdAt,
		LastUpdateAt:   entity.updatedAt,
		IdempotencyKey: entity.idempotencyKey,
		Configuration: domain.WebhookConfiguration{
			Url: entity.webhookURL,
		},
		Metadata: metadata,
	}, nil
}

func applyAttemptToEntity(entity *messageEntity, attempt domain.Attempt) {
	switch attempt := attempt.(type) {
	case domain.SuccessAttempt:
		attemptType := "success"
		phoneID := string(attempt.PhoneId)
		entity.lastAttemptType = &attemptType
		entity.lastAttemptPhoneID = &phoneID
		entity.lastAttemptCount = &attempt.AttemptCount
	case domain.FailedAttempt:
		attemptType := "failure"
		phoneID := string(attempt.PhoneId)
		entity.lastAttemptType = &attemptType
		entity.lastAttemptPhoneID = &phoneID
		entity.lastAttemptCount = &attempt.AttemptCount
		entity.lastAttemptFailureReason = &attempt.Reason
	}
}

func entityToAttempt(entity messageEntity) domain.Attempt {
	if entity.lastAttemptType == nil || entity.lastAttemptCount == nil {
		return nil
	}

	phoneID := domain.PhoneId("")
	if entity.lastAttemptPhoneID != nil {
		phoneID = domain.PhoneId(*entity.lastAttemptPhoneID)
	}

	switch *entity.lastAttemptType {
	case "success":
		return domain.SuccessAttempt{
			AttemptCount: *entity.lastAttemptCount,
			PhoneId:      phoneID,
		}
	case "failure":
		reason := ""
		if entity.lastAttemptFailureReason != nil {
			reason = *entity.lastAttemptFailureReason
		}
		return domain.FailedAttempt{
			AttemptCount: *entity.lastAttemptCount,
			PhoneId:      phoneID,
			Reason:       reason,
		}
	default:
		return nil
	}
}
