package user_account

import (
	"cloud.google.com/go/firestore"
	"context"
	"time"
)

type FirestoreUserAccountRepository struct {
	context    context.Context
	store      *firestore.Client
	collection string
}

type UserAccountJsonEntity struct {
	Phone       string    `firestore:"phone"`
	ApiKey      string    `firestore:"apiKey"`
	IsSuspended bool      `firestore:"isSuspended"`
	CreatedAt   time.Time `firestore:"createdAt"`
}

func NewFirestoreUserAccountRepository(ctx context.Context, store *firestore.Client, collection string) FirestoreUserAccountRepository {
	return FirestoreUserAccountRepository{context: ctx, store: store, collection: collection}
}

func (i FirestoreUserAccountRepository) Save(account UserAccount) (bool, error) {
	accountEntity := UserAccountJsonEntity{
		Phone:       account.Phone,
		ApiKey:      string(account.ApiKey),
		CreatedAt:   account.CreatedAt,
		IsSuspended: account.IsSuspended,
	}
	if _, err := i.store.Collection(i.collection).Doc(string(account.Id)).Set(i.context, accountEntity); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (i FirestoreUserAccountRepository) FindById(accountId AccountId) *UserAccount {
	if account, err := i.store.Collection(i.collection).Doc(string(accountId)).Get(i.context); err != nil {
		var entity UserAccountJsonEntity
		if err := account.DataTo(&entity); err == nil {
			return &UserAccount{
				Id:          AccountId(account.Ref.ID),
				ApiKey:      ApiKey(entity.ApiKey),
				Phone:       entity.Phone,
				IsSuspended: entity.IsSuspended,
				CreatedAt:   entity.CreatedAt,
			}
		}
	}
	return nil
}

func (i FirestoreUserAccountRepository) FindByApiKey(apiKey ApiKey) *UserAccount {
	accounts, err := i.store.Collection(i.collection).Where("apiKey", "==", apiKey).Limit(1).Documents(i.context).GetAll()
	if err != nil {
		return nil
	}
	if len(accounts) == 0 {
		return nil
	}
	var entity UserAccountJsonEntity
	if err := accounts[0].DataTo(&entity); err != nil {
		return nil
	}
	return &UserAccount{
		Id:          AccountId(accounts[0].Ref.ID),
		ApiKey:      ApiKey(entity.ApiKey),
		Phone:       entity.Phone,
		IsSuspended: entity.IsSuspended,
		CreatedAt:   entity.CreatedAt,
	}
}
