package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AccountID string
type ApiKey string

type UserAccount struct {
	Id          AccountID
	Phone       string
	ApiKey      ApiKey
	IsSuspended bool
	CreatedAt   time.Time
}

type UserAccountRepository interface {
	Save(UserAccount) (bool, error)
	FindById(AccountID) *UserAccount
	FindByApiKey(ApiKey) *UserAccount
}

func NewUserAccount(phone string) UserAccount {
	// TODO: validate Phone number
	return UserAccount{
		Id:          AccountID(uuid.NewString()),
		ApiKey:      ApiKey(strings.ReplaceAll(uuid.NewString(), "-", "")),
		Phone:       phone,
		IsSuspended: false,
		CreatedAt:   time.Now(),
	}
}

func (account UserAccount) String() string {
	return fmt.Sprintf(
		"SmsId %s, Phone %s, ApiKey %s, IsSuspende %t, CreatedAt %s",
		string(account.Id),
		account.Phone,
		string(account.ApiKey),
		account.IsSuspended,
		account.CreatedAt,
	)
}
