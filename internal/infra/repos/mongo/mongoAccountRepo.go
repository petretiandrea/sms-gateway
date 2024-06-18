package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sms-gateway/internal/domain"
	"time"
)

type MongoUserAccountRepository struct {
	context    context.Context
	collection *mongo.Collection
}

type UserAccountJsonEntity struct {
	Id          string    `bson:"_id"`
	Phone       string    `bson:"phone"`
	ApiKey      string    `bson:"apiKey"`
	IsSuspended bool      `bson:"isSuspended"`
	CreatedAt   time.Time `bson:"createdAt"`
}

func NewMongoUserAccountRepository(ctx context.Context, collection *mongo.Collection) MongoUserAccountRepository {
	return MongoUserAccountRepository{context: ctx, collection: collection}
}

func (i MongoUserAccountRepository) Save(account domain.UserAccount) (bool, error) {
	accountEntity := UserAccountJsonEntity{
		Id:          string(account.Id),
		Phone:       account.Phone,
		ApiKey:      string(account.ApiKey),
		CreatedAt:   account.CreatedAt,
		IsSuspended: account.IsSuspended,
	}
	if _, err := i.collection.UpdateByID(
		i.context,
		string(account.Id),
		bson.D{{"$set", accountEntity}},
		options.Update().SetUpsert(true),
	); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (i MongoUserAccountRepository) FindById(accountId domain.AccountID) *domain.UserAccount {
	var entity *UserAccountJsonEntity
	if err := i.collection.FindOne(i.context, bson.D{primitive.E{Key: "_id", Value: accountId}}).Decode(&entity); err == nil {
		return &domain.UserAccount{
			Id:          domain.AccountID(entity.Id),
			ApiKey:      domain.ApiKey(entity.ApiKey),
			Phone:       entity.Phone,
			IsSuspended: entity.IsSuspended,
			CreatedAt:   entity.CreatedAt,
		}
	}
	return nil
}

func (i MongoUserAccountRepository) FindByApiKey(apiKey domain.ApiKey) *domain.UserAccount {
	var entity UserAccountJsonEntity
	if err := i.collection.FindOne(i.context, bson.D{primitive.E{Key: "apiKey", Value: apiKey}}).Decode(&entity); err == nil {
		return &domain.UserAccount{
			Id:          domain.AccountID(entity.Id),
			ApiKey:      domain.ApiKey(entity.ApiKey),
			Phone:       entity.Phone,
			IsSuspended: entity.IsSuspended,
			CreatedAt:   entity.CreatedAt,
		}
	} else {
		return nil
	}
}

var (
	_ domain.UserAccountRepository = (*MongoUserAccountRepository)(nil)
)
