package device_gateway

import (
	"context"
	"sms-gateway/internal/sms"
	"sms-gateway/internal/user_account"
	"time"

	"cloud.google.com/go/firestore"
)

const phoneKey = "phone"

type FirestorePhoneRepository struct {
	context    context.Context
	store      *firestore.Client
	collection string
}

type PhoneJsonEntity struct {
	Phone     string    `firestore:"phone"`
	Account   string    `firestore:"account"`
	FCMToken  string    `firestore:"fcmToken"`
	CreatedAt time.Time `firestore:"createdAt"`
	UpdatedAt time.Time `firestore:"updatedAt"`
}

func NewFirestorePhoneRepository(ctx context.Context, store *firestore.Client, collection string) FirestorePhoneRepository {
	return FirestorePhoneRepository{context: ctx, store: store, collection: collection}
}

func (repo *FirestorePhoneRepository) Save(phone Phone) (*Phone, error) {
	entity := phone.toEntity()
	if _, err := repo.store.Collection(repo.collection).Doc(string(phone.Id)).Set(repo.context, entity); err != nil {
		return nil, err
	}
	return &phone, nil
}

func (repo *FirestorePhoneRepository) FindById(id PhoneId) *Phone {
	if snapshot, err := repo.store.Collection(repo.collection).Doc(string(id)).Get(repo.context); err != nil {
		return nil
	} else {
		var message PhoneJsonEntity
		if err := snapshot.DataTo(&message); err != nil {
			return nil
		} else {
			return message.toMessage(snapshot.Ref.ID)
		}
	}
}

func (repo *FirestorePhoneRepository) FindByPhoneNumber(number sms.PhoneNumber) *Phone {
	snapshot, err := repo.store.
		Collection(repo.collection).
		Where(phoneKey, "==", number.Number).
		Limit(1).
		Documents(repo.context).
		GetAll()
	if err != nil {
		return nil
	}
	if len(snapshot) == 0 {
		return nil
	}
	var message PhoneJsonEntity
	if err := snapshot[0].DataTo(&message); err != nil {
		return nil
	}
	return message.toMessage(snapshot[0].Ref.ID)
}

func (repo *FirestorePhoneRepository) Delete(id PhoneId) bool {
	if _, err := repo.store.Doc(string(id)).Delete(repo.context); err != nil {
		return false
	}
	return true
}

func (phone *Phone) toEntity() PhoneJsonEntity {
	return PhoneJsonEntity{
		Phone:     phone.Phone.Number,
		Account:   string(phone.UserId),
		FCMToken:  string(phone.Token),
		CreatedAt: phone.CreatedAt,
		UpdatedAt: phone.UpdatedAt,
	}
}

func (entity *PhoneJsonEntity) toMessage(id string) *Phone {
	return &Phone{
		Id:        PhoneId(id),
		Phone:     sms.PhoneNumber{Number: entity.Phone},
		UserId:    user_account.AccountId(entity.Account),
		Token:     FCMToken(entity.FCMToken),
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
