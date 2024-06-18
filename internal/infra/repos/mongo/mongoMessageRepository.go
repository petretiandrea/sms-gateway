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

func (r MongoMessageRepository) Find(params domain.QueryParams) ([]domain.Sms, error) {
	filter := bson.M{}
	if params.From != "" {
		filter["from"] = params.From
	}
	if params.IsSent != nil {
		filter["isSent"] = *params.IsSent
	}
	find, err := r.collection.Find(
		r.context,
		filter,
	)
	if err != nil {
		return nil, err
	}
	var messages []MongoMessageEntity
	err = find.All(r.context, &messages)
	if err != nil {
		return nil, err
	}
	var domainMessages []domain.Sms
	for _, message := range messages {
		domainMessages = append(domainMessages, *message.ToMessage(message.Id))
	}
	return domainMessages, nil
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

var (
	_ domain.Repository = (*MongoMessageRepository)(nil)
)
