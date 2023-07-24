package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"time"
)

type SmsEntityResponse struct {
	Id           string    `json:"id" binding:"required"`
	Content      string    `json:"content" binding:"required"`
	From         string    `json:"from" binding:"required"`
	To           string    `json:"to" binding:"required"`
	UserId       string    `json:"owner" binding:"required"`
	IsSent       bool      `json:"isSent" binding:"required"`
	SendAttempts int       `json:"sendAttempts" binding:"required"`
	CreatedAt    time.Time `json:"createdAt" binding:"required"`
}

type SendSmsRequest struct {
	Content string `json:"content" binding:"required"`
	From    string `json:"from" binding:"required"`
	To      string `json:"to" binding:"required"`
}

type SmsApiController struct {
	Account application.UserAccountService
	Sms     application.SmsService
}

func (controller *SmsApiController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/sms")
	router.Use(NewApiKeyMiddleware(controller.Account))
	router.POST("/", controller.sendSms)
}

func (controller *SmsApiController) sendSms(context *gin.Context) {
	var sendRequest SendSmsRequest
	user := context.MustGet("user").(domain.UserAccount)
	idempotencyKey := context.GetHeader("Idempotency-Key")
	if err := context.BindJSON(&sendRequest); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	if idempotencyKey == "" {
		context.JSON(http.StatusBadRequest, "Missing mandatory idempotency key")
		return
	}
	sendCommand := application.SendSmsCommand{
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
	if message := controller.Sms.GetSMS(domain.SmsId(context.Param("messageId"))); message == nil {
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
