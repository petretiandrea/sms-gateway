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

func NewMessageFirestoreRepository(ctx context.Context, store *firestore.Client, collection string) FirestoreMessageRepository {
	return FirestoreMessageRepository{context: ctx, store: store, collection: collection}
}

func (repo FirestoreMessageRepository) Save(message domain.Sms) (*domain.Sms, error) {
	entity := smsMapToEntity(message)

	if _, err := repo.store.Collection(repo.collection).Doc(string(message.Id)).Set(repo.context, entity); err != nil {
		return nil, err
	}
	return &message, nil
}

func (repo FirestoreMessageRepository) FindById(id domain.SmsId) *domain.Sms {
	if snapshot, err := repo.store.Collection(repo.collection).Doc(string(id)).Get(repo.context); err != nil {
		return nil
	} else {
		var message MessageFirestoreEntity
		if err := snapshot.DataTo(&message); err != nil {
			return nil
		} else {
			return message.ToMessage(snapshot.Ref.ID)
		}
	}
}

func (repo FirestoreMessageRepository) FindExisting(idempotencyKey string) *domain.Sms {
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
	var message MessageFirestoreEntity
	if err := snapshot[0].DataTo(&message); err != nil {
		return nil
	}
	return message.ToMessage(snapshot[0].Ref.ID)
}

var _ domain.Repository = (*FirestoreMessageRepository)(nil)
