package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/generated/openapi"
)

type PhoneApiController struct {
	Phone   application.PhoneService
	Account application.UserAccountService
}

func (p PhoneApiController) GetPhoneById(c *gin.Context) {
	phoneId := c.Param("phoneId")
	if device, err := p.Phone.GetPhoneById(domain.PhoneId(phoneId)); device != nil {
		c.JSONP(http.StatusOK, openapi.PhoneEntityResponse{
			Id:        string(device.Id),
			Phone:     device.Phone.Number,
			Account:   string(device.UserId),
			FcmToken:  string(device.Token),
			CreatedAt: device.CreatedAt,
			UpdatedAt: device.UpdatedAt,
		})
	} else {
		if err == nil {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.JSON(http.StatusBadRequest, err)
		}
	}
}

func (p PhoneApiController) PhonePost(context *gin.Context) {
	var sendRequest openapi.RegisterPhoneRequestDto
	user := context.MustGet("user").(domain.UserAccount)
	if err := context.BindJSON(&sendRequest); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}

	if device, err := p.Phone.RegisterPhone(
		domain.PhoneNumber{Number: sendRequest.Phone},
		user.Id,
	); err == nil {
		context.JSONP(http.StatusCreated, device.Id)
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
}

func (p PhoneApiController) UpdateFcmToken(c *gin.Context) {
	var request openapi.UpdatePhoneFirebaseTokenDto
	phoneId := c.Param("phoneId")
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if device, err := p.Phone.UpdateFCMToken(
		domain.PhoneId(phoneId),
		domain.FCMToken(request.Token),
	); err == nil {
		c.JSONP(http.StatusOK, device.Id)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

var (
	_ openapi.PhoneAPI = (*PhoneApiController)(nil)
)
