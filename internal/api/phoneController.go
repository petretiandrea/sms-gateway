package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"time"
)

type RegisterPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type UpdateFCMRequest struct {
	Token string `json:"token" binding:"required"`
}

type PhoneResponse struct {
	Id        string    `json:"id" binding:"required"`
	Phone     string    `json:"phone" binding:"required"`
	Account   string    `json:"account" binding:"required"`
	FCMToken  string    `json:"fcmToken" binding:"required"`
	CreatedAt time.Time `json:"createdAt" binding:"required"`
	UpdatedAt time.Time `json:"updatedAt" binding:"required"`
}

type PhoneApiController struct {
	Phone   application.PhoneService
	Account application.UserAccountService
}

func (controller *PhoneApiController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/phone")
	router.Use(NewApiKeyMiddleware(controller.Account))
	router.POST("/", controller.registerPhone)
	router.PUT("/:phoneId", controller.updateToken)
	router.GET("/:phoneId", controller.getPhone)
}

func (controller *PhoneApiController) registerPhone(context *gin.Context) {
	var sendRequest RegisterPhoneRequest
	user := context.MustGet("user").(domain.UserAccount)
	if err := context.BindJSON(&sendRequest); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}

	if device, err := controller.Phone.RegisterPhone(
		domain.PhoneNumber{Number: sendRequest.Phone},
		user.Id,
	); err == nil {
		context.JSONP(http.StatusCreated, device.Id)
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
}

func (controller *PhoneApiController) updateToken(context *gin.Context) {
	var request UpdateFCMRequest
	phoneId := context.Param("phoneId")
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	if device, err := controller.Phone.UpdateFCMToken(
		domain.PhoneId(phoneId),
		domain.FCMToken(request.Token),
	); err == nil {
		context.JSONP(http.StatusOK, device.Id)
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
}

func (controller *PhoneApiController) getPhone(context *gin.Context) {
	phoneId := context.Param("phoneId")
	if device, err := controller.Phone.GetPhoneById(domain.PhoneId(phoneId)); device != nil {
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
