package domain

import (
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
	Save(phone Phone) (*Phone, error)
	FindById(id PhoneId) *Phone
	FindByPhoneNumber(number PhoneNumber) *Phone
	Delete(id PhoneId) bool
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
