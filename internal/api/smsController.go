package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/generated/openapi"
)

type SmsApiController struct {
	Account           application.UserAccountService
	Sms               application.SmsService
	MessageRepository domain.Repository
}

func (s SmsApiController) GetMessages(c *gin.Context) {
	var queryParams domain.QueryParams
	err := c.ShouldBindQuery(&queryParams)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	messages, err := s.MessageRepository.Find(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var responses []openapi.SmsEntityResponse
	for _, message := range messages {
		responses = append(responses, openapi.SmsEntityResponse{
			Id:          string(message.Id),
			To:          message.To,
			From:        message.From.Number,
			Content:     message.Content,
			Owner:       string(message.UserId),
			CreatedAt:   message.CreatedAt,
			IsSent:      message.IsSent,
			LastAttempt: lastAttemptToDto(message.LastAttempt),
		})
	}
	c.JSON(http.StatusOK, openapi.GetMessages200Response{
		Messages: responses,
	})
	return
}

func (s SmsApiController) GetSmsById(c *gin.Context) {
	if message := s.Sms.GetSMS(domain.SmsId(c.Param("smsId"))); message != nil {
		c.JSONP(http.StatusOK, openapi.SmsEntityResponse{
			Id:          string(message.Id),
			To:          message.To,
			From:        message.From.Number,
			Content:     message.Content,
			Owner:       string(message.UserId),
			CreatedAt:   message.CreatedAt,
			IsSent:      message.IsSent,
			LastAttempt: lastAttemptToDto(message.LastAttempt),
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
	sendCommand := application.CreateMessageCommand{
		From:           sendRequest.From,
		To:             sendRequest.To,
		Content:        sendRequest.Content,
		IdempotencyKey: idempotencyKey,
		Account:        user,
		WebhookUrl:     sendRequest.Webhook.Url,
		Metadata:       sendRequest.Metadata,
	}
	if createMessage, err := s.Sms.SendSMS(sendCommand); err == nil && createMessage != nil {
		c.JSONP(http.StatusCreated, openapi.SmsEntityResponse{
			Id:          string(createMessage.Id),
			To:          createMessage.To,
			From:        createMessage.From.Number,
			Content:     createMessage.Content,
			Owner:       string(createMessage.UserId),
			CreatedAt:   createMessage.CreatedAt,
			IsSent:      createMessage.IsSent,
			LastAttempt: lastAttemptToDto(createMessage.LastAttempt),
		})
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func lastAttemptToDto(attempt domain.Attempt) *openapi.SmsEntityResponseLastAttempt {
	if _, ok := attempt.(domain.SuccessAttempt); ok {
		return &openapi.SmsEntityResponseLastAttempt{
			Type:         "success",
			AttemptCount: attempt.AttemptNumber(),
		}
	} else if failure, ok := attempt.(domain.FailedAttempt); ok {
		return &openapi.SmsEntityResponseLastAttempt{
			Type:         "failure",
			Reason:       failure.Reason,
			AttemptCount: attempt.AttemptNumber(),
		}
	}
	return nil
}

var (
	_ openapi.SmsAPI = (*SmsApiController)(nil)
)
