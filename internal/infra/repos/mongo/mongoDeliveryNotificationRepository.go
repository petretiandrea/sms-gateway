package mongo

import (
	"context"
	"sms-gateway/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDeliveryNotificationRepository struct {
	collection *mongo.Collection
}

func NewMongoDeliveryNotificationRepository(collection *mongo.Collection) MongoDeliveryNotificationRepository {
	return MongoDeliveryNotificationRepository{
		collection: collection,
	}
}

func (m MongoDeliveryNotificationRepository) Save(ctx context.Context, config domain.DeliveryNotificationConfig) (bool, error) {
	doc := document{
		AccountId:  config.AccountId,
		Enabled:    config.Enabled,
		WebhookURL: config.WebhookURL,
	}
	if _, err := m.collection.UpdateByID(
		ctx,
		doc.AccountId,
		bson.D{{"$set", doc}},
		options.Update().SetUpsert(true),
	); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (m MongoDeliveryNotificationRepository) FindById(ctx context.Context, id domain.AccountID) *domain.DeliveryNotificationConfig {
	var doc *document
	err := m.collection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&doc)
	if err != nil {
		return nil
	}
	return &domain.DeliveryNotificationConfig{
		AccountId:  doc.AccountId,
		Enabled:    doc.Enabled,
		WebhookURL: doc.WebhookURL,
	}
}

type document struct {
	AccountId  domain.AccountID `bson:"accountId"`
	WebhookURL string           `bson:"webhookURL"`
	Enabled    bool             `bson:"enabled"`
}

var _ domain.DeliveryNotificationConfigRepository = (*MongoDeliveryNotificationRepository)(nil)
