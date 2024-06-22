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
		responses = append(responses, smsToResponseEntity(&message))
	}
	c.JSON(http.StatusOK, openapi.GetMessages200Response{
		Messages: responses,
	})
	return
}

func (s SmsApiController) GetSmsById(c *gin.Context) {
	if message := s.Sms.GetSMS(domain.SmsId(c.Param("smsId"))); message != nil {
		c.JSONP(http.StatusOK, smsToResponseEntity(message))
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
		c.JSONP(http.StatusCreated, smsToResponseEntity(createMessage))
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

func smsToResponseEntity(sms *domain.Sms) openapi.SmsEntityResponse {
	return openapi.SmsEntityResponse{
		Id:          string(sms.Id),
		To:          sms.To,
		From:        sms.From.Number,
		Content:     sms.Content,
		Owner:       string(sms.UserId),
		CreatedAt:   sms.CreatedAt,
		IsSent:      sms.IsSent,
		LastAttempt: lastAttemptToDto(sms.LastAttempt),
		UpdatedAt:   sms.LastUpdateAt,
	}
}

var (
	_ openapi.SmsAPI = (*SmsApiController)(nil)
)
