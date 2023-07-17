package api

import (
	"net/http"
	"sms-gateway/internal/user_account"
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
	CreateUserAccountUseCase user_account.UserAccountService
}

func (controller *UserAccountController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/account")
	router.POST("/", controller.CreateUserAccount)
	router.GET("/:accountId", controller.GetUserAccount)
}

func (controller *UserAccountController) CreateUserAccount(context *gin.Context) {
	var createRequest CreateAccountRequest
	if err := context.BindJSON(&createRequest); err == nil {
		accountRequest := user_account.CreateNewAccountParams{Phone: createRequest.PhoneNumber}
		if account, err := controller.CreateUserAccountUseCase.CreateNewAccount(accountRequest); err == nil {
			context.JSONP(http.StatusCreated, UserAccountResponse{
				AccountId:   string(account.Id),
				PhoneNumber: account.Phone,
				ApiKey:      string(account.ApiKey),
			})
		} else {
			context.JSON(http.StatusBadRequest, err)
		}
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
	return
}

func (controller *UserAccountController) GetUserAccount(context *gin.Context) {
	accountId := context.Param("accountId")
	if account := controller.CreateUserAccountUseCase.GetUserAccount(user_account.AccountId(accountId)); account != nil {
		context.JSONP(http.StatusOK, UserAccountResponse{
			AccountId:   string(account.Id),
			PhoneNumber: account.Phone,
			IsActive:    !account.IsSuspended,
			CreateAt:    account.CreatedAt,
		})
	} else {
		context.AbortWithStatus(http.StatusNotFound)
	}
	return
}
