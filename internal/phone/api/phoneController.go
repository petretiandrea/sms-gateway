package api

import (
	"net/http"
	"sms-gateway/internal/account"
	"sms-gateway/internal/adapter"
	"sms-gateway/internal/phone"
	"sms-gateway/internal/sms"

	"github.com/gin-gonic/gin"
)

type PhoneApiController struct {
	Phone   phone.Service
	Account account.UserAccountService
}

func (controller *PhoneApiController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/phone")
	router.Use(adapter.NewApiKeyMiddleware(controller.Account))
	router.POST("/", controller.RegisterPhone)
	router.PUT("/:phoneId", controller.UpdateToken)
	router.GET("/:phoneId", controller.GetPhone)
}

func (controller *PhoneApiController) RegisterPhone(context *gin.Context) {
	var sendRequest RegisterPhoneRequest
	user := context.MustGet("user").(account.UserAccount)
	if err := context.BindJSON(&sendRequest); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}

	if device, err := controller.Phone.RegisterPhone(
		sms.PhoneNumber{Number: sendRequest.Phone},
		user.Id,
	); err == nil {
		context.JSONP(http.StatusCreated, device.Id)
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
}

func (controller *PhoneApiController) UpdateToken(context *gin.Context) {
	var request UpdateFCMRequest
	phoneId := context.Param("phoneId")
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	if device, err := controller.Phone.UpdateFCMToken(
		phone.PhoneId(phoneId),
		phone.FCMToken(request.Token),
	); err == nil {
		context.JSONP(http.StatusOK, device.Id)
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
}

func (controller *PhoneApiController) GetPhone(context *gin.Context) {
	phoneId := context.Param("phoneId")
	if device, err := controller.Phone.GetPhoneById(phone.PhoneId(phoneId)); device != nil {
		context.JSONP(http.StatusOK, PhoneResponse{
			Id:        string(device.Id),
			Phone:     device.Phone.Number,
			Account:   string(device.UserId),
			FCMToken:  string(device.Token),
			CreatedAt: device.CreatedAt,
			UpdatedAt: device.UpdatedAt,
		})
	} else {
		if err == nil {
			context.AbortWithStatus(http.StatusNotFound)
		} else {
			context.JSON(http.StatusBadRequest, err)
		}
	}
}
