package mongo

import (
	"context"
	"sms-gateway/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const idempotencyKeyName = "idempotencyKey"

type MongoMessageRepository struct {
	collection *mongo.Collection
}

func NewMongoMessageRepository(collection *mongo.Collection) MongoMessageRepository {
	return MongoMessageRepository{collection: collection}
}

func (r MongoMessageRepository) Find(ctx context.Context, params domain.QueryParams) ([]domain.Sms, error) {
	filter := bson.M{}
	if params.From != "" {
		filter["from"] = params.From
	}
	if params.IsSent != nil {
		filter["isSent"] = *params.IsSent
	}
	find, err := r.collection.Find(
		ctx,
		filter,
	)
	if err != nil {
		return nil, err
	}
	var messages []MongoMessageEntity
	err = find.All(ctx, &messages)
	if err != nil {
		return nil, err
	}
	var domainMessages []domain.Sms
	for _, message := range messages {
		domainMessages = append(domainMessages, *message.ToMessage(message.Id))
	}
	return domainMessages, nil
}

func (r MongoMessageRepository) Save(ctx context.Context, message *domain.Sms) (*domain.Sms, error) {
	entity := smsMapToEntity(*message)
	if _, err := r.collection.UpdateByID(
		ctx,
		string(message.Id),
		bson.D{{"$set", entity}}, options.Update().SetUpsert(true),
	); err != nil {
		return nil, err
	}
	return message, nil
}

func (r MongoMessageRepository) FindById(ctx context.Context, id domain.SmsId) *domain.Sms {
	// if err := i.collection.FindOne(i.context, bson.D{primitive.E{Key: "_id", Value: accountId}}).Decode(&entity); err == nil {
	var message MongoMessageEntity
	if err := r.collection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&message); err == nil {
		return message.ToMessage(message.Id)
	}
	return nil
}

func (r MongoMessageRepository) FindExisting(ctx context.Context, idempotencyKey string) *domain.Sms {
	// if err := i.collection.FindOne(i.context, bson.D{primitive.E{Key: "apiKey", Value: apiKey}}).Decode(&entity); err == nil {
	var message MongoMessageEntity
	if err := r.collection.FindOne(ctx, bson.D{primitive.E{Key: idempotencyKeyName, Value: idempotencyKey}}).Decode(&message); err == nil {
		return message.ToMessage(message.Id)
	} else {
		return nil
	}
}

var (
	_ domain.Repository = (*MongoMessageRepository)(nil)
)
