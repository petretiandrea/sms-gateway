package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/sms"
	"sms-gateway/internal/user_account"
)

type SmsApiController struct {
	Account user_account.UserAccountService
	Sms     sms.Service
}

func (controller *SmsApiController) RegisterRoutes(gin *gin.Engine) {
	router := gin.Group("/sms")
	router.POST("/", controller.SendSms)
}

func (controller *SmsApiController) SendSms(context *gin.Context) {
	var sendRequest SendSmsRequest
	var apiKey = context.GetHeader("Api-Key")
	var idempotencyKey = context.GetHeader("Idempotency-Key")
	if err := context.BindJSON(&sendRequest); err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}
	if idempotencyKey == "" {
		context.JSON(http.StatusBadRequest, "Missing mandatory idempotency key")
		return
	}
	if account := controller.Account.GetUserAccountByApiKey(user_account.ApiKey(apiKey)); account == nil {
		context.JSON(http.StatusBadRequest, "Cannot find Account based on given api key")
		return
	} else {
		sendCommand := sms.SendSmsCommand{
			From:           sendRequest.From,
			To:             sendRequest.To,
			Content:        sendRequest.Content,
			IdempotencyKey: idempotencyKey,
			Account:        *account,
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
}
