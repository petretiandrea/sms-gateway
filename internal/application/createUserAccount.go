package application

import (
	"go.uber.org/zap"
	"sms-gateway/internal/domain"
)

type UserAccountService struct {
	repo domain.UserAccountRepository
	log  *zap.Logger
}

func NewUserAccountService(repository domain.UserAccountRepository) UserAccountService {
	return UserAccountService{repo: repository}
}

type CreateNewAccountParams struct {
	Phone string
}

func (service *UserAccountService) CreateNewAccount(params CreateNewAccountParams) (*domain.UserAccount, error) {
	account := domain.NewUserAccount(params.Phone)
	if _, err := service.repo.Save(account); err == nil {
		return &account, nil
	} else {
		return nil, err
	}
}

func (service *UserAccountService) GetUserAccount(id domain.AccountID) *domain.UserAccount {
	return service.repo.FindById(id)
}

func (service *UserAccountService) GetUserAccountByApiKey(apiKey domain.ApiKey) *domain.UserAccount {
	return service.repo.FindByApiKey(apiKey)
}

func (service *UserAccountService) GetUserAccountApiKey(id domain.AccountID) *domain.ApiKey {
	if account := service.repo.FindById(id); account != nil {
		return &account.ApiKey
	} else {
		return nil
	}
}
