package repos

import (
	"cloud.google.com/go/firestore"
	"context"
	"sms-gateway/internal/domain"
)

const idempotencyKeyName = "idempotencyKey"

type FirestoreMessageRepository struct {
	context    context.Context
	store      *firestore.Client
	collection string
}

type MessageJsonEntity struct {
	From           string `firestore:"from"`
	To             string `firestore:"to"`
	Content        string `firestore:"content"`
	IsSent         bool   `firestore:"isSent"`
	SendAttempts   uint8  `firestore:"sendAttempts"`
	Owner          string `firestore:"owner"`
	IdempotencyKey string `firestore:"idempotencyKey"`
}

func NewMessageFirestoreRepository(ctx context.Context, store *firestore.Client, collection string) FirestoreMessageRepository {
	return FirestoreMessageRepository{context: ctx, store: store, collection: collection}
}

func (repo *FirestoreMessageRepository) Save(message domain.Sms) (*domain.Sms, error) {
	entity := smsMapToEntity(message)
	if _, err := repo.store.Collection(repo.collection).Doc(string(message.Id)).Set(repo.context, entity); err != nil {
		return nil, err
	}
	return &message, nil
}

func (repo *FirestoreMessageRepository) FindById(id domain.SmsId) *domain.Sms {
	if snapshot, err := repo.store.Collection(repo.collection).Doc(string(id)).Get(repo.context); err != nil {
		return nil
	} else {
		var message MessageJsonEntity
		if err := snapshot.DataTo(&message); err != nil {
			return nil
		} else {
			return message.toMessage(snapshot.Ref.ID)
		}
	}
}

func (repo *FirestoreMessageRepository) FindExisting(idempotencyKey string) *domain.Sms {
	snapshot, err := repo.store.
		Collection(repo.collection).
		Where(idempotencyKeyName, "==", idempotencyKey).
		Limit(1).
		Documents(repo.context).
		GetAll()
	if err != nil {
		return nil
	}
	if len(snapshot) == 0 {
		return nil
	}
	var message MessageJsonEntity
	if err := snapshot[0].DataTo(&message); err != nil {
		return nil
	}
	return message.toMessage(snapshot[0].Ref.ID)
}

func smsMapToEntity(message domain.Sms) MessageJsonEntity {
	return MessageJsonEntity{
		From:           message.From.Number,
		To:             message.To,
		Content:        message.Content,
		IsSent:         message.IsSent,
		SendAttempts:   uint8(message.SendAttempts),
		Owner:          string(message.UserId),
		IdempotencyKey: message.IdempotencyKey,
	}
}

func (entity *MessageJsonEntity) toMessage(id string) *domain.Sms {
	return &domain.Sms{
		Id:             domain.SmsId(id),
		From:           domain.PhoneNumber{Number: entity.From},
		To:             entity.To,
		IsSent:         entity.IsSent,
		Content:        entity.Content,
		UserId:         domain.AccountID(entity.Owner),
		IdempotencyKey: entity.IdempotencyKey,
	}
}
