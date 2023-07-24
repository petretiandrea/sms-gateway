package application

import (
	"errors"
	"go.uber.org/zap"
	"sms-gateway/internal/domain"
)

type PhoneService struct {
	repo domain.PhoneRepository
	log  *zap.Logger
}

func NewPhoneService(repo domain.PhoneRepository) PhoneService {
	return PhoneService{repo: repo}
}

func (service *PhoneService) RegisterPhone(phoneNumber domain.PhoneNumber, userId domain.AccountID) (*domain.Phone, error) {
	if existingPhone := service.repo.FindByPhoneNumber(phoneNumber); existingPhone != nil {
		return existingPhone, nil
	} else {
		newPhone := domain.NewPhone(phoneNumber, userId, "")
		return service.repo.Save(newPhone)
	}
}

func (service *PhoneService) UpdateFCMToken(id domain.PhoneId, token domain.FCMToken) (*domain.Phone, error) {
	if phone := service.repo.FindById(id); phone != nil {
		phone.UpdateFCMToken(token)
		return service.repo.Save(*phone)
	}
	return nil, errors.New("no phone with given id found")
}

func (service *PhoneService) RemovePhone(id domain.PhoneId) bool {
	return service.repo.Delete(id)
}

func (service *PhoneService) GetPhoneById(id domain.PhoneId) (*domain.Phone, error) {
	return service.repo.FindById(id), nil
}

func (service *PhoneService) GetPhoneByNumber(number domain.PhoneNumber) (*domain.Phone, error) {
	return service.repo.FindByPhoneNumber(number), nil
}
