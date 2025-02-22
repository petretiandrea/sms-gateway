package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PhoneId string
type FCMToken string

type Phone struct {
	Id        PhoneId
	Phone     PhoneNumber
	UserId    AccountID
	Token     FCMToken
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PhoneRepository interface {
	Save(ctx context.Context, phone Phone) (*Phone, error)
	FindById(ctx context.Context, id PhoneId) *Phone
	FindByPhoneNumber(ctx context.Context, number PhoneNumber) *Phone
	Delete(ctx context.Context, id PhoneId) bool
}

func NewPhone(phone PhoneNumber, accountId AccountID, token FCMToken) Phone {
	return Phone{
		Id:        PhoneId(uuid.NewString()),
		Phone:     phone,
		UserId:    accountId,
		Token:     token,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (phone *Phone) UpdateFCMToken(newToken FCMToken) {
	phone.Token = newToken
	phone.UpdatedAt = time.Now()
}
