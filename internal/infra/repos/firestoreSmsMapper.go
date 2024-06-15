package repos

import (
	"sms-gateway/internal/domain"
	"time"
)

type MessageFirestoreEntity struct {
	From           string            `firestore:"from"`
	To             string            `firestore:"to"`
	Content        string            `firestore:"content"`
	IsSent         bool              `firestore:"isSent"`
	LastAttempt    AttemptDocument   `firestore:"lastAttempt"`
	Owner          string            `firestore:"owner"`
	IdempotencyKey string            `firestore:"idempotencyKey"`
	CreatedAt      time.Time         `firestore:"createdAt"`
	UpdatedAt      time.Time         `firestore:"updatedAt"`
	Metadata       map[string]string `firestore:"additionalData"`
}

type AttemptDocument struct {
	Type          string `firestore:"type"`
	PhoneId       string `firestore:"phoneId"`
	AttemptCount  int32  `firestore:"attemptCount"`
	FailureReason string `firestore:"failureReason"`
}

func smsMapToEntity(message domain.Sms) MessageFirestoreEntity {
	return MessageFirestoreEntity{
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

func (entity *MessageFirestoreEntity) ToMessage(id string) *domain.Sms {
	return &domain.Sms{
		Id:             domain.SmsId(id),
		From:           domain.PhoneNumber{Number: entity.From},
		To:             entity.To,
		IsSent:         entity.IsSent,
		LastAttempt:    mapAttemptToModel(entity.LastAttempt),
		Content:        entity.Content,
		UserId:         domain.AccountID(entity.Owner),
		IdempotencyKey: entity.IdempotencyKey,
		CreatedAt:      entity.CreatedAt,
		LastUpdateAt:   entity.UpdatedAt,
		Metadata:       entity.Metadata,
	}
}

func mapAttemptToDocument(attempt domain.Attempt) AttemptDocument {
	if success, ok := attempt.(domain.SuccessAttempt); ok {
		return AttemptDocument{Type: "success", AttemptCount: success.AttemptCount, PhoneId: string(success.PhoneId)}
	} else if failure, ok := attempt.(domain.FailedAttempt); ok {
		return AttemptDocument{Type: "failure", FailureReason: failure.Reason, AttemptCount: failure.AttemptCount, PhoneId: string(success.PhoneId)}
	}
	return AttemptDocument{}
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
