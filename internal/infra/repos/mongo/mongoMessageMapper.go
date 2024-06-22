package mongo

import (
	"sms-gateway/internal/domain"
	"time"
)

type MongoMessageEntity struct {
	Id             string            `bson:"_id"`
	From           string            `bson:"from"`
	To             string            `bson:"to"`
	Content        string            `bson:"content"`
	IsSent         bool              `bson:"isSent"`
	LastAttempt    *AttemptDocument  `bson:"lastAttempt,omitempty"`
	Owner          string            `bson:"owner"`
	IdempotencyKey string            `bson:"idempotencyKey"`
	CreatedAt      time.Time         `bson:"createdAt"`
	UpdatedAt      time.Time         `bson:"updatedAt"`
	Metadata       map[string]string `bson:"additionalData"`
}

type AttemptDocument struct {
	Type          string `bson:"type"`
	PhoneId       string `bson:"phoneId"`
	AttemptCount  int32  `bson:"attemptCount"`
	FailureReason string `bson:"failureReason"`
}

func smsMapToEntity(message domain.Sms) MongoMessageEntity {
	return MongoMessageEntity{
		Id:             string(message.Id),
		From:           message.From.Number,
		To:             message.To,
		Content:        message.Content,
		IsSent:         message.IsSent,
		LastAttempt:    mapAttemptToDocument(message.LastAttempt),
		Owner:          string(message.UserId),
		IdempotencyKey: message.IdempotencyKey,
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      time.Now(),
		Metadata:       message.Metadata,
	}
}

func (entity *MongoMessageEntity) ToMessage(id string) *domain.Sms {
	return &domain.Sms{
		Id:             domain.SmsId(id),
		From:           domain.PhoneNumber{Number: entity.From},
		To:             entity.To,
		IsSent:         entity.IsSent,
		LastAttempt:    mapAttemptToModel(*entity.LastAttempt),
		Content:        entity.Content,
		UserId:         domain.AccountID(entity.Owner),
		IdempotencyKey: entity.IdempotencyKey,
		CreatedAt:      entity.CreatedAt,
		LastUpdateAt:   entity.UpdatedAt,
		Metadata:       entity.Metadata,
	}
}

func mapAttemptToDocument(attempt domain.Attempt) *AttemptDocument {
	if success, ok := attempt.(domain.SuccessAttempt); ok {
		return &AttemptDocument{Type: "success", AttemptCount: success.AttemptCount, PhoneId: string(success.PhoneId)}
	} else if failure, ok := attempt.(domain.FailedAttempt); ok {
		return &AttemptDocument{Type: "failure", FailureReason: failure.Reason, AttemptCount: failure.AttemptCount, PhoneId: string(success.PhoneId)}
	}
	return nil
}

func mapAttemptToModel(attempt AttemptDocument) domain.Attempt {
	switch attempt.Type {
	case "success":
		return domain.SuccessAttempt{
			AttemptCount: attempt.AttemptCount,
			PhoneId:      domain.PhoneId(attempt.PhoneId),
		}
	case "failure":
		return domain.FailedAttempt{
			AttemptCount: attempt.AttemptCount,
			Reason:       attempt.FailureReason,
			PhoneId:      domain.PhoneId(attempt.PhoneId),
		}
	}
	return nil
}
