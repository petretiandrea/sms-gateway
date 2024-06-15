package application

import (
	"sms-gateway/internal/domain"
	"sms-gateway/internal/infra"
)

type SmsService struct {
	phone        PhoneService
	repo         domain.Repository
	notification infra.FirebasePushNotification
}

type CreateMessageCommand struct {
	Content        string
	To             string
	From           string
	Account        domain.UserAccount
	IdempotencyKey string
	WebhookUrl     string
	Metadata       map[string]string
}

func NewSmsService(repo domain.Repository, phoneService PhoneService, pushService infra.FirebasePushNotification) SmsService {
	return SmsService{repo: repo, phone: phoneService, notification: pushService}
}

func (service *SmsService) SendSMS(params CreateMessageCommand) (*domain.Sms, error) {
	if message := service.repo.FindExisting(params.IdempotencyKey); message != nil {
		return message, nil
	} else {
		// retrieve phoneAccount associated
		var metadata map[string]string
		if params.Metadata == nil {
			metadata = make(map[string]string)
		} else {
			metadata = params.Metadata
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
		_, err := service.repo.Save(message)
		if err != nil {
			return nil, err
		}
		phoneAccount, err := service.phone.GetPhoneByNumber(message.From)
		if err != nil {
			return nil, err
		}
		if err := service.notification.Send(message, string(phoneAccount.Token)); err != nil {
			return nil, err
		}
		return &message, nil
	}
}

func (service *SmsService) GetSMS(id domain.SmsId) *domain.Sms {
	return service.repo.FindById(id)
}
