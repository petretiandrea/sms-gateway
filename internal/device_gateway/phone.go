package device_gateway

import (
	"github.com/google/uuid"
	"sms-gateway/internal/sms"
	"sms-gateway/internal/user_account"
	"time"
)

type PhoneId string
type FCMToken string

type Phone struct {
	Id        PhoneId
	Phone     sms.PhoneNumber
	UserId    user_account.AccountId
	Token     FCMToken
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PhoneRepository interface {
	Save(phone Phone) (*Phone, error)
	FindById(id PhoneId) *Phone
	FindByPhoneNumber(number sms.PhoneNumber) *Phone
	Delete(id PhoneId) bool
}

func NewPhone(phone sms.PhoneNumber, accountId user_account.AccountId, token FCMToken) Phone {
	return Phone{
		Id:        PhoneId(uuid.NewString()),
		Phone:     phone,
		UserId:    accountId,
		Token:     token,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (phone *Phone) updateFCMToken(newToken FCMToken) {
	phone.Token = newToken
	phone.UpdatedAt = time.Now()
}
