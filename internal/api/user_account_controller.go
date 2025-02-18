package api

import (
	"context"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
)

type UserAccountController struct {
	CreateUserAccountUseCase application.UserAccountService
}

// RegisterAccount implements StrictServerInterface
func (c *UserAccountController) RegisterAccount(
	ctx context.Context, 
	request RegisterAccountRequestObject,
) (RegisterAccountResponseObject, error) {
	accountRequest := application.CreateNewAccountParams{
		Phone: request.Body.PhoneNumber,
	}
	if newAccount, err := c.CreateUserAccountUseCase.CreateNewAccount(accountRequest); err == nil {
		return RegisterAccount201JSONResponse{
			AccountId: string(newAccount.Id),
			PhoneNumber: newAccount.Phone,
			ApiKey: string(newAccount.ApiKey),
		}, nil
	} else {
		return RegisterAccount400Response{}, nil
	}
}

// GetAccountById implements StrictServerInterface
func (c *UserAccountController) GetAccountById(
	ctx context.Context, 
	request GetAccountByIdRequestObject,
) (GetAccountByIdResponseObject, error) {
	accountId := domain.AccountID(request.AccountId.String())
	if foundAccount := c.CreateUserAccountUseCase.GetUserAccount(accountId); foundAccount != nil {
		return GetAccountById200JSONResponse{
			AccountId: string(foundAccount.Id),
			PhoneNumber: foundAccount.Phone,
			IsActive: !foundAccount.IsSuspended,
			CreateAt: &foundAccount.CreatedAt,
		}, nil
	} else {
		return GetAccountById404Response{}, nil
	}
}