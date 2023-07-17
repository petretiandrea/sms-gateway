package phone

import (
	"sms-gateway/internal/account"
	"sms-gateway/internal/sms"
	"time"

	"github.com/google/uuid"
)

type PhoneId string
type FCMToken string

type Phone struct {
	Id        PhoneId
	Phone     sms.PhoneNumber
	UserId    account.AccountID
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

func NewPhone(phone sms.PhoneNumber, accountId account.AccountID, token FCMToken) Phone {
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
