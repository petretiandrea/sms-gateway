package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/generated/openapi"
)

type SmsApiController struct {
	Account application.UserAccountService
	Sms     application.SmsService
}

func (s SmsApiController) GetSmsById(c *gin.Context) {
	if message := s.Sms.GetSMS(domain.SmsId(c.Param("smsId"))); message == nil {
		c.JSON(http.StatusOK, openapi.SmsEntityResponse{
			Id:           string(message.Id),
			To:           message.To,
			From:         message.From.Number,
			Content:      message.Content,
			Owner:        string(message.UserId),
			CreatedAt:    message.CreatedAt,
			IsSent:       message.IsSent,
			SendAttempts: int32(message.SendAttempts),
		})
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
	return
}

func (s SmsApiController) SendSms(c *gin.Context) {
	var sendRequest openapi.SendSmsRequest
	user := c.MustGet("user").(domain.UserAccount)
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if err := c.BindJSON(&sendRequest); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, "Missing mandatory idempotency key")
		return
	}
	sendCommand := application.SendSmsCommand{
		From:           sendRequest.From,
		To:             sendRequest.To,
		Content:        sendRequest.Content,
		IdempotencyKey: idempotencyKey,
		Account:        user,
		Metadata:       sendRequest.Metadata,
	}
	if createMessage, err := s.Sms.SendSMS(sendCommand); err == nil {
		c.JSONP(http.StatusCreated, openapi.SmsEntityResponse{
			Id:           string(createMessage.Id),
			To:           createMessage.To,
			From:         createMessage.From.Number,
			Content:      createMessage.Content,
			Owner:        string(createMessage.UserId),
			CreatedAt:    createMessage.CreatedAt,
			IsSent:       createMessage.IsSent,
			SendAttempts: int32(createMessage.SendAttempts),
		})
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

var (
	_ openapi.SmsAPI = (*SmsApiController)(nil)
)
