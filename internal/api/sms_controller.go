package api

import (
	"context"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"

	"github.com/google/uuid"
)

type SmsApiController struct {
	Account           application.UserAccountService
	Sms               application.SmsService
	MessageRepository domain.Repository
}

// GetMessages implements StrictServerInterface.
func (c *SmsApiController) GetMessages(
	ctx context.Context, 
	request GetMessagesRequestObject,
) (GetMessagesResponseObject, error) {
	queryParams := domain.QueryParams {
		From: *request.Params.From,
		IsSent: request.Params.IsSent,
	}
	messages, err := c.MessageRepository.Find(queryParams)
	if err != nil {
		return nil, err
	}
	var responses []SmsEntityResponse
	for _, message := range messages {
		responses = append(responses, smsToResponseEntity(&message))
	}
	return GetMessages200JSONResponse{Messages: &responses}, nil
}

// GetSmsById implements StrictServerInterface.
func (c *SmsApiController) GetSmsById(ctx context.Context, request GetSmsByIdRequestObject) (GetSmsByIdResponseObject, error) {
	if message := c.Sms.GetSMS(domain.SmsId(request.SmsId.String())); message != nil {
		return GetSmsById200JSONResponse(smsToResponseEntity(message)), nil
	}
	return GetSmsById404Response{}, nil
}


// SendSms implements StrictServerInterface.
func (c *SmsApiController) SendSms(
	ctx context.Context, 
	request SendSmsRequestObject,
) (SendSmsResponseObject, error) {
	user := ctx.Value("user").(domain.UserAccount)
	idempotencyKey := request.Params.IdempotencyKey
	if idempotencyKey == nil {
		return SendSms400Response{}, nil
	}
	metadata := (*map[string]string)(request.Body.Metadata)
	var webhookUrl *string
	if (request.Body.Webhook != nil) {
		webhookUrl = request.Body.Webhook.Url
	}
	sendCommand := application.CreateMessageCommand{
		From:           request.Body.From,
		To:             request.Body.To,
		Content:        request.Body.Content,
		IdempotencyKey: *idempotencyKey,
		Account:        user,
		WebhookUrl:     webhookUrl,
		Metadata:       metadata,
	}
	if createMessage, err := c.Sms.SendSMS(sendCommand); err == nil && createMessage != nil {
		return SendSms201JSONResponse(smsToResponseEntity(createMessage)), nil
	} else {
		return SendSms400Response{}, nil
	}
}

func smsToResponseEntity(sms *domain.Sms) SmsEntityResponse {
	return SmsEntityResponse{
		Id:          uuid.MustParse(string(sms.Id)),
		To:          sms.To,
		From:        sms.From.Number,
		Content:     sms.Content,
		Owner:       uuid.MustParse(string(sms.UserId)),
		CreatedAt:   sms.CreatedAt,
		IsSent:      sms.IsSent,
		LastAttempt: lastAttemptToDto(sms.LastAttempt),
		UpdatedAt:   sms.LastUpdateAt,
	}
}

func lastAttemptToDto(attempt domain.Attempt) *SmsEntityResponse_LastAttempt {
	var attemptResponse = SmsEntityResponse_LastAttempt{
		AttemptCount: int(attempt.AttemptNumber()),
	}
	if _, ok := attempt.(domain.SuccessAttempt); ok {
		attemptResponse.FromSuccessfulAttempt(SuccessfulAttempt{})
		return &attemptResponse
	} else if failure, ok := attempt.(domain.FailedAttempt); ok {
		attemptResponse.FromFailedAttempt(FailedAttempt{
			Reason: &failure.Reason,
		})
		return &attemptResponse
	}
	return nil
}