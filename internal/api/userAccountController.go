package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/generated/openapi"
)

type UserAccountController struct {
	CreateUserAccountUseCase application.UserAccountService
}

func (controller UserAccountController) AccountPost(context *gin.Context) {
	var createRequest openapi.CreateAccountRequestDto
	if err := context.BindJSON(&createRequest); err == nil {
		accountRequest := application.CreateNewAccountParams{Phone: createRequest.PhoneNumber}
		if newAccount, err := controller.CreateUserAccountUseCase.CreateNewAccount(accountRequest); err == nil {
			context.JSONP(http.StatusCreated, openapi.AccountEntityDto{
				AccountId:   string(newAccount.Id),
				PhoneNumber: newAccount.Phone,
				ApiKey:      string(newAccount.ApiKey),
			})
		} else {
			context.JSON(http.StatusBadRequest, err)
		}
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
	return
}

func (controller UserAccountController) GetAccountById(context *gin.Context) {
	accountId := context.Param("accountId")
	if foundAccount := controller.CreateUserAccountUseCase.GetUserAccount(domain.AccountID(accountId)); foundAccount != nil {
		context.JSONP(http.StatusOK, openapi.AccountEntityDto{
			AccountId:   string(foundAccount.Id),
			PhoneNumber: foundAccount.Phone,
			IsActive:    !foundAccount.IsSuspended,
			CreateAt:    foundAccount.CreatedAt,
		})
	} else {
		context.AbortWithStatus(http.StatusNotFound)
	}
	return
}

var (
	_ openapi.AccountAPI = (*UserAccountController)(nil)
)
