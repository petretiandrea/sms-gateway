package application

import (
	"context"
	"sms-gateway/internal/domain"

	"go.uber.org/zap"
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

func (service *UserAccountService) CreateNewAccount(ctx context.Context, params CreateNewAccountParams) (*domain.UserAccount, error) {
	account := domain.NewUserAccount(params.Phone)
	if _, err := service.repo.Save(ctx, account); err == nil {
		return &account, nil
	} else {
		return nil, err
	}
}

func (service *UserAccountService) GetUserAccount(ctx context.Context, id domain.AccountID) *domain.UserAccount {
	return service.repo.FindById(ctx, id)
}

func (service *UserAccountService) GetUserAccountByApiKey(ctx context.Context, apiKey domain.ApiKey) *domain.UserAccount {
	return service.repo.FindByApiKey(ctx, apiKey)
}

func (service *UserAccountService) GetUserAccountApiKey(ctx context.Context, id domain.AccountID) *domain.ApiKey {
	if account := service.repo.FindById(ctx, id); account != nil {
		return &account.ApiKey
	} else {
		return nil
	}
}
