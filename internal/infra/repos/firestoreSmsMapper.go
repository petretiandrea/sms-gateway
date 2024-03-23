package repos

import (
	"sms-gateway/internal/domain"
	"time"
)

type MessageFirestoreEntity struct {
	From           string    `firestore:"from"`
	To             string    `firestore:"to"`
	Content        string    `firestore:"content"`
	IsSent         bool      `firestore:"isSent"`
	SendAttempts   uint8     `firestore:"sendAttempts"`
	Owner          string    `firestore:"owner"`
	IdempotencyKey string    `firestore:"idempotencyKey"`
	CreatedAt      time.Time `firestore:"createdAt"`
	UpdatedAt      time.Time `firestore:"updatedAt"`
}

func smsMapToEntity(message domain.Sms) MessageFirestoreEntity {
	return MessageFirestoreEntity{
		From:           message.From.Number,
		To:             message.To,
		Content:        message.Content,
		IsSent:         message.IsSent,
		SendAttempts:   uint8(message.SendAttempts),
		Owner:          string(message.UserId),
		IdempotencyKey: message.IdempotencyKey,
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      time.Now(),
	}
}

func (entity *MessageFirestoreEntity) ToMessage(id string) *domain.Sms {
	return &domain.Sms{
		Id:             domain.SmsId(id),
		From:           domain.PhoneNumber{Number: entity.From},
		To:             entity.To,
		IsSent:         entity.IsSent,
		Content:        entity.Content,
		UserId:         domain.AccountID(entity.Owner),
		IdempotencyKey: entity.IdempotencyKey,
		CreatedAt:      entity.CreatedAt,
		LastUpdateAt:   entity.UpdatedAt,
	}
}
