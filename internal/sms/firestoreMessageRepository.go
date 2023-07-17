package sms

import (
	"cloud.google.com/go/firestore"
	"context"
	"sms-gateway/internal/user_account"
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

func (repo *FirestoreMessageRepository) Save(message Message) (*Message, error) {
	entity := message.toEntity()
	if _, err := repo.store.Collection(repo.collection).Doc(string(message.Id)).Set(repo.context, entity); err != nil {
		return nil, err
	}
	return &message, nil
}

func (repo *FirestoreMessageRepository) FindById(id MessageId) *Message {
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

func (repo *FirestoreMessageRepository) FindExisting(idempotencyKey string) *Message {
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

func (message *Message) toEntity() MessageJsonEntity {
	return MessageJsonEntity{
		From:           message.From.Number,
		To:             message.To,
		Content:        message.Content,
		IsSent:         message.IsSent,
		SendAttempts:   uint8(message.SendAttempts),
		Owner:          string(message.UserId),
		IdempotencyKey: message.idempotencyKey,
	}
}

func (entity *MessageJsonEntity) toMessage(id string) *Message {
	return &Message{
		Id:             MessageId(id),
		From:           PhoneNumber{Number: entity.From},
		To:             entity.To,
		IsSent:         entity.IsSent,
		Content:        entity.Content,
		UserId:         user_account.AccountId(entity.Owner),
		idempotencyKey: entity.IdempotencyKey,
	}
}
