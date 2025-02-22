package application

import (
	"context"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/infra"

	"github.com/pkg/errors"
)

type SmsService struct {
	phone                 PhoneService
	repo                  domain.Repository
	notification          infra.FirebasePushNotification
	messageFeedController domain.MessageChangeFeedController
}

type CreateMessageCommand struct {
	Content        string
	To             string
	From           string
	Account        domain.UserAccount
	IdempotencyKey string
	WebhookUrl     *string
	Metadata       *map[string]string
}

func NewSmsService(
	repo domain.Repository,
	phoneService PhoneService,
	pushService infra.FirebasePushNotification,
	messageFeedController domain.MessageChangeFeedController,
) SmsService {
	return SmsService{repo: repo, phone: phoneService, notification: pushService, messageFeedController: messageFeedController}
}

func (service *SmsService) SendSMS(ctx context.Context, params CreateMessageCommand) (*domain.Sms, error) {
	if message := service.repo.FindExisting(ctx, params.IdempotencyKey); message != nil {
		return message, nil
	} else {
		// retrieve phoneAccount associated
		var metadata map[string]string
		if params.Metadata == nil {
			metadata = make(map[string]string)
		} else {
			metadata = *params.Metadata
		}
		message := domain.CreateNewSMS(
			params.Account.Id,
			domain.PhoneNumber{Number: params.From},
			domain.PhoneNumber{Number: params.To},
			params.Content,
			params.IdempotencyKey,
			metadata,
			domain.WebhookConfiguration{Url: params.WebhookUrl},
		)
		phoneAccount, err := service.phone.GetPhoneByNumber(ctx, message.From)
		if err != nil {
			return nil, err
		}
		if phoneAccount == nil {
			return nil, nil
		}
		_, err = service.repo.Save(ctx, message)
		if err != nil {
			return nil, err
		}
		if err := service.notification.Send(ctx, message, string(phoneAccount.Token)); err != nil {
			return nil, err
		}
		return &message, nil
	}
}

func (service *SmsService) GetSMS(ctx context.Context, id domain.SmsId) *domain.Sms {
	return service.repo.FindById(ctx, id)
}

func (service *SmsService) RegisterAttempt(
	ctx context.Context,
	id domain.SmsId,
	accountID domain.AccountID,
	attempt domain.Attempt,
) (*domain.Sms, error) {
	sms := service.repo.FindById(ctx, id)
	if sms == nil {
		return nil, nil
	}
	if sms.UserId == accountID {
		sms.RegisterAttempt(attempt)
		save, err := service.repo.Save(ctx, *sms)
		if err != nil {
			return nil, err
		} else {
			service.messageFeedController.Add(*save)
			return save, nil
		}
	} else {
		return nil, errors.Wrapf(
			domain.ErrorNotMessageOwner,
			"Not owner of sms [%s]",
			sms.Id,
		)
	}
}
