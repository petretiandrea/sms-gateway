package repos

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sms-gateway/internal/domain"
)

type MongoDeliveryNotificationRepository struct {
	ctx   context.Context
	mongo *mongo.Collection
}

const collectionName = "deliveryconfigs"

func NewMongoDeliveryNotificationRepository(ctx context.Context, mongo *mongo.Client, databaseName string) MongoDeliveryNotificationRepository {
	return MongoDeliveryNotificationRepository{
		ctx:   ctx,
		mongo: mongo.Database(databaseName).Collection(collectionName),
	}
}

func (m MongoDeliveryNotificationRepository) Save(config domain.DeliveryNotificationConfig) (bool, error) {
	doc := document{
		AccountId:  config.AccountId,
		Enabled:    config.Enabled,
		WebhookURL: config.WebhookURL,
	}
	if _, err := m.mongo.UpdateByID(
		m.ctx,
		doc.AccountId,
		bson.D{{"$set", doc}},
		options.Update().SetUpsert(true),
	); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (m MongoDeliveryNotificationRepository) FindById(id domain.AccountID) *domain.DeliveryNotificationConfig {
	var doc *document
	err := m.mongo.FindOne(m.ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&doc)
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
