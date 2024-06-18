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

const phoneKey = "phone"

type MongoPhoneRepository struct {
	context    context.Context
	collection *mongo.Collection
}

type PhoneJsonEntity struct {
	Id        string    `bson:"_id"`
	Phone     string    `bson:"phone"`
	Account   string    `bson:"account"`
	FCMToken  string    `bson:"fcmToken"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func NewMongoPhoneRepository(ctx context.Context, collection *mongo.Collection) MongoPhoneRepository {
	return MongoPhoneRepository{context: ctx, collection: collection}
}

func (r *MongoPhoneRepository) Save(phone domain.Phone) (*domain.Phone, error) {
	entity := phoneToEntity(phone)
	if _, err := r.collection.UpdateByID(
		r.context,
		string(phone.Id),
		bson.D{{"$set", entity}}, options.Update().SetUpsert(true),
	); err != nil {
		return nil, err
	}
	return &phone, nil

}

func (r *MongoPhoneRepository) FindById(id domain.PhoneId) *domain.Phone {
	var message PhoneJsonEntity
	if err := r.collection.FindOne(r.context, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&message); err == nil {
		return message.toMessage(message.Id)
	}
	return nil
}

func (r *MongoPhoneRepository) FindByPhoneNumber(number domain.PhoneNumber) *domain.Phone {
	var message PhoneJsonEntity
	if err := r.collection.FindOne(r.context, bson.D{primitive.E{Key: phoneKey, Value: number.Number}}).Decode(&message); err == nil {
		return message.toMessage(message.Id)
	}
	return nil
}

func (r *MongoPhoneRepository) Delete(id domain.PhoneId) bool {
	if _, err := r.collection.DeleteOne(r.context, bson.D{primitive.E{Key: "_id", Value: id}}); err != nil {
		return false
	}
	return true
}

func phoneToEntity(phone domain.Phone) PhoneJsonEntity {
	return PhoneJsonEntity{
		Id:        string(phone.Id),
		Phone:     phone.Phone.Number,
		Account:   string(phone.UserId),
		FCMToken:  string(phone.Token),
		CreatedAt: phone.CreatedAt,
		UpdatedAt: phone.UpdatedAt,
	}
}

func (entity *PhoneJsonEntity) toMessage(id string) *domain.Phone {
	return &domain.Phone{
		Id:        domain.PhoneId(id),
		Phone:     domain.PhoneNumber{Number: entity.Phone},
		UserId:    domain.AccountID(entity.Account),
		Token:     domain.FCMToken(entity.FCMToken),
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
