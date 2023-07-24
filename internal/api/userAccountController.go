package api

import (
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateAccountRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

type UserAccountResponse struct {
	AccountId   string    `json:"accountId"`
	PhoneNumber string    `json:"phoneNumber"`
	ApiKey      string    `json:"apiKey,omitempty"`
	IsActive    bool      `json:"isActive"`
	CreateAt    time.Time `json:"createAt"`
}

type UserAccountController struct {
	CreateUserAccountUseCase application.UserAccountService
}

func (controller *UserAccountController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/account")
	router.POST("/", controller.createUserAccount)
	router.GET("/:accountId", controller.getUserAccount)
}

func (controller *UserAccountController) createUserAccount(context *gin.Context) {
	var createRequest CreateAccountRequest
	if err := context.BindJSON(&createRequest); err == nil {
		accountRequest := application.CreateNewAccountParams{Phone: createRequest.PhoneNumber}
		if newAccount, err := controller.CreateUserAccountUseCase.CreateNewAccount(accountRequest); err == nil {
			context.JSONP(http.StatusCreated, UserAccountResponse{
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

func (controller *UserAccountController) getUserAccount(context *gin.Context) {
	accountId := context.Param("accountId")
	if foundAccount := controller.CreateUserAccountUseCase.GetUserAccount(domain.AccountID(accountId)); foundAccount != nil {
		context.JSONP(http.StatusOK, UserAccountResponse{
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
