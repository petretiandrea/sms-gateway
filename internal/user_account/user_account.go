package user_account

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type AccountId string
type ApiKey string

type UserAccount struct {
	Id          AccountId
	Phone       string
	ApiKey      ApiKey
	IsSuspended bool
	CreatedAt   time.Time
}

type UserAccountRepository interface {
	Save(UserAccount) (bool, error)
	FindById(AccountId) *UserAccount
	FindByApiKey(ApiKey) *UserAccount
}

func NewUserAccount(phone string) UserAccount {
	// TODO: validate Phone number
	return UserAccount{
		Id:          AccountId(uuid.NewString()),
		ApiKey:      ApiKey(strings.ReplaceAll(uuid.NewString(), "-", "")),
		Phone:       phone,
		IsSuspended: false,
		CreatedAt:   time.Now(),
	}
}

func (account UserAccount) String() string {
	return fmt.Sprintf(
		"Id %s, Phone %s, ApiKey %s, IsSuspende %t, CreatedAt %s",
		string(account.Id),
		account.Phone,
		string(account.ApiKey),
		account.IsSuspended,
		account.CreatedAt,
	)
}
