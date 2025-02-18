package api

import (
	"context"
	"errors"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
)

type AttemptController struct {
	SmsService application.SmsService
}

// ReportMessageStatus implements StrictServerInterface.
func (c *AttemptController) ReportMessageStatus(
	ctx context.Context, 
	request ReportMessageStatusRequestObject,
) (ReportMessageStatusResponseObject, error) {
	user := ctx.Value("user").(domain.UserAccount)
	// TODO: replace by sending it to a queue\
	if attempt := reportRequestToAttemptDomain(request); attempt == nil {
		return nil, errors.New("Invalid attempt type")
	} else {
		if sms, err := c.SmsService.RegisterAttempt(
			domain.SmsId(request.Body.MessageId.String()),
			user.Id,
			attempt,
		); err != nil {
			return ReportMessageStatus500Response{}, nil
		} else if sms == nil {
			return ReportMessageStatus400Response{}, nil
		} else {
			return ReportMessageStatus202Response{}, nil
		}
	}
}

func reportRequestToAttemptDomain(request ReportMessageStatusRequestObject) domain.Attempt {
	if attemptRaw, err := request.Body.Result.ValueByDiscriminator(); err != nil {
		switch attempt := attemptRaw.(type) {
		case SuccessfulAttempt:
			// Handle SuccessfulAttempt
			return domain.SuccessAttempt{
				AttemptCount: int32(request.Body.Attempt),
				PhoneId:      domain.PhoneId(request.Body.PhoneId.String()),
			}
		case FailedAttempt:
			return domain.FailedAttempt{
				AttemptCount: int32(request.Body.Attempt),
				Reason:       *attempt.Reason,
				PhoneId:      domain.PhoneId(request.Body.PhoneId.String()),
			}
		default:
			return nil
		}
	}
	return nil
}
