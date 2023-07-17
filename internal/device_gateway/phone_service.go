package device_gateway

import (
	"errors"
	"sms-gateway/internal/sms"
	"sms-gateway/internal/user_account"
)

type Service struct {
	repo PhoneRepository
}

func NewPhoneService(repo PhoneRepository) Service {
	return Service{repo: repo}
}

func (service *Service) RegisterPhone(phoneNumber sms.PhoneNumber, userId user_account.AccountId) (*Phone, error) {
	if existingPhone := service.repo.FindByPhoneNumber(phoneNumber); existingPhone != nil {
		return existingPhone, nil
	} else {
		newPhone := NewPhone(phoneNumber, userId, "")
		return service.repo.Save(newPhone)
	}
}

func (service *Service) UpdateFCMToken(id PhoneId, token FCMToken) (*Phone, error) {
	if phone := service.repo.FindById(id); phone != nil {
		phone.updateFCMToken(token)
		return service.repo.Save(*phone)
	}
	return nil, errors.New("no phone with given id found")
}

func (service *Service) RemovePhone(id PhoneId) bool {
	return service.repo.Delete(id)
}

func (service *Service) GetPhoneById(id PhoneId) (*Phone, error) {
	return service.repo.FindById(id), nil
}
