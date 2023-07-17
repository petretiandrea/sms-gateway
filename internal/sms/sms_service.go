package sms

import (
	"sms-gateway/internal/user_account"
)

type Service struct {
	account user_account.UserAccountRepository
	repo    MessageRepository
}

type SendSmsCommand struct {
	Content        string
	To             string
	From           string
	Account        user_account.UserAccount
	IdempotencyKey string
}

func NewSmsService(repo MessageRepository) Service {
	return Service{repo: repo}
}

/*
TODO: TODO: 1. retrieve user phone settings, for specific api key
and FROM (the setting could be FCM Token associated to a device which
owns the From phone number.
*/
func (service *Service) SendSMS(params SendSmsCommand) (*Message, error) {
	if message := service.repo.FindExisting(params.IdempotencyKey); message != nil {
		return message, nil
	} else {
		message := CreateNewSMS(
			params.Account.Id,
			PhoneNumber{Number: params.To},
			PhoneNumber{Number: params.From},
			params.Content,
			params.IdempotencyKey,
		)
		return service.repo.Save(message)
	}
}

func (service *Service) GetSMS(id MessageId) *Message {
	return service.repo.FindById(id)
}
