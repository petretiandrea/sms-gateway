package phone

import (
	"errors"
	"sms-gateway/internal/account"
	"sms-gateway/internal/sms"
)

type Service struct {
	repo PhoneRepository
}

func NewPhoneService(repo PhoneRepository) Service {
	return Service{repo: repo}
}

func (service *Service) RegisterPhone(phoneNumber sms.PhoneNumber, userId account.AccountID) (*Phone, error) {
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
