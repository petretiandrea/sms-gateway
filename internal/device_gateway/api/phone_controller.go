package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/device_gateway"
	"sms-gateway/internal/sms"
	"sms-gateway/internal/user_account"
)

type PhoneApiController struct {
	Phone   device_gateway.Service
	Account user_account.UserAccountService
}

func (controller *PhoneApiController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/phone")
	router.POST("/", controller.RegisterPhone)
	router.PUT("/:phoneId", controller.UpdateToken)
	router.GET("/:phoneId", controller.GetPhone)
}

func (controller *PhoneApiController) RegisterPhone(context *gin.Context) {
	var sendRequest RegisterPhoneRequest
	var apiKey = context.GetHeader("Api-Key")
	if err := context.BindJSON(&sendRequest); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	if account := controller.Account.GetUserAccountByApiKey(user_account.ApiKey(apiKey)); account == nil {
		context.JSON(http.StatusBadRequest, "Cannot find Account based on given api key")
		return
	} else {
		if phone, err := controller.Phone.RegisterPhone(
			sms.PhoneNumber{Number: sendRequest.Phone},
			account.Id,
		); err == nil {
			context.JSONP(http.StatusCreated, phone.Id)
		} else {
			context.JSON(http.StatusBadRequest, err)
		}
	}
}

func (controller *PhoneApiController) UpdateToken(context *gin.Context) {
	var request UpdateFCMRequest
	apiKey := context.GetHeader("Api-Key")
	phoneId := context.Param("phoneId")
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	if account := controller.Account.GetUserAccountByApiKey(user_account.ApiKey(apiKey)); account == nil {
		context.JSON(http.StatusBadRequest, "Cannot find Account based on given api key")
		return
	} else {
		if phone, err := controller.Phone.UpdateFCMToken(
			device_gateway.PhoneId(phoneId),
			device_gateway.FCMToken(request.Token),
		); err == nil {
			context.JSONP(http.StatusOK, phone.Id)
		} else {
			context.JSON(http.StatusBadRequest, err)
		}
	}
}

func (controller *PhoneApiController) GetPhone(context *gin.Context) {
	apiKey := context.GetHeader("Api-Key")
	phoneId := context.Param("phoneId")
	if account := controller.Account.GetUserAccountByApiKey(user_account.ApiKey(apiKey)); account == nil {
		context.JSON(http.StatusBadRequest, "Cannot find Account based on given api key")
		return
	} else {
		if phone, err := controller.Phone.GetPhoneById(device_gateway.PhoneId(phoneId)); phone != nil {
			context.JSONP(http.StatusOK, PhoneResponse{
				Id:        string(phone.Id),
				Phone:     phone.Phone.Number,
				Account:   string(phone.UserId),
				FCMToken:  string(phone.Token),
				CreatedAt: phone.CreatedAt,
				UpdatedAt: phone.UpdatedAt,
			})
		} else {
			if err == nil {
				context.AbortWithStatus(http.StatusNotFound)
			} else {
				context.JSON(http.StatusBadRequest, err)
			}
		}
	}
}
