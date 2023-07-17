package api

import (
	"net/http"
	"sms-gateway/internal/account"
	"sms-gateway/internal/adapter"
	"sms-gateway/internal/sms"

	"github.com/gin-gonic/gin"
)

type SmsApiController struct {
	Account account.UserAccountService
	Sms     sms.Service
}

func (controller *SmsApiController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/sms")
	router.Use(adapter.NewApiKeyMiddleware(controller.Account))
	router.POST("/", controller.SendSms)
}

func (controller *SmsApiController) SendSms(context *gin.Context) {
	var sendRequest SendSmsRequest
	user := context.MustGet("user").(account.UserAccount)
	idempotencyKey := context.GetHeader("Idempotency-Key")
	if err := context.BindJSON(&sendRequest); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	if idempotencyKey == "" {
		context.JSON(http.StatusBadRequest, "Missing mandatory idempotency key")
		return
	}
	sendCommand := sms.SendSmsCommand{
		From:           sendRequest.From,
		To:             sendRequest.To,
		Content:        sendRequest.Content,
		IdempotencyKey: idempotencyKey,
		Account:        user,
	}
	if createMessage, err := controller.Sms.SendSMS(sendCommand); err == nil {
		context.JSONP(http.StatusCreated, SmsEntityResponse{
			Id:           string(createMessage.Id),
			To:           createMessage.To,
			From:         createMessage.From.Number,
			Content:      createMessage.Content,
			UserId:       string(createMessage.UserId),
			CreatedAt:    createMessage.CreatedAt,
			IsSent:       createMessage.IsSent,
			SendAttempts: createMessage.SendAttempts,
		})
	} else {
		context.JSON(http.StatusBadRequest, err)
	}
}

func (controller *SmsApiController) GetSms(context *gin.Context) {
	if message := controller.Sms.GetSMS(sms.MessageId(context.Param("messageId"))); message == nil {
		context.JSON(http.StatusOK, SmsEntityResponse{
			Id:           string(message.Id),
			To:           message.To,
			From:         message.From.Number,
			Content:      message.Content,
			UserId:       string(message.UserId),
			CreatedAt:    message.CreatedAt,
			IsSent:       message.IsSent,
			SendAttempts: message.SendAttempts,
		})
		return
	}

	context.AbortWithStatus(http.StatusNotFound)
	return
}
