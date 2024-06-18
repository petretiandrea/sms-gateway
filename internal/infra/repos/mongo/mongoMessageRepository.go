package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sms-gateway/internal/domain"
)

const idempotencyKeyName = "idempotencyKey"

type MongoMessageRepository struct {
	context    context.Context
	collection *mongo.Collection
}

func NewMongoMessageRepository(ctx context.Context, collection *mongo.Collection) MongoMessageRepository {
	return MongoMessageRepository{context: ctx, collection: collection}
}

func (r MongoMessageRepository) Save(message domain.Sms) (*domain.Sms, error) {
	entity := smsMapToEntity(message)
	if _, err := r.collection.UpdateByID(
		r.context,
		string(message.Id),
		bson.D{{"$set", entity}}, options.Update().SetUpsert(true),
	); err != nil {
		return nil, err
	}
	return &message, nil
}

func (r MongoMessageRepository) FindById(id domain.SmsId) *domain.Sms {
	// if err := i.collection.FindOne(i.context, bson.D{primitive.E{Key: "_id", Value: accountId}}).Decode(&entity); err == nil {
	var message MongoMessageEntity
	if err := r.collection.FindOne(r.context, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&message); err == nil {
		return message.ToMessage(message.Id)
	}
	return nil
}

func (r MongoMessageRepository) FindExisting(idempotencyKey string) *domain.Sms {
	// if err := i.collection.FindOne(i.context, bson.D{primitive.E{Key: "apiKey", Value: apiKey}}).Decode(&entity); err == nil {
	var message MongoMessageEntity
	if err := r.collection.FindOne(r.context, bson.D{primitive.E{Key: idempotencyKeyName, Value: idempotencyKey}}).Decode(&message); err == nil {
		return message.ToMessage(message.Id)
	} else {
		return nil
	}
}

var _ domain.Repository = (*MongoMessageRepository)(nil)
