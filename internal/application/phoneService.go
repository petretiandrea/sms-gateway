package application

import (
	"context"
	"errors"
	"sms-gateway/internal/domain"

	"go.uber.org/zap"
)

type PhoneService struct {
	repo domain.PhoneRepository
	log  *zap.Logger
}

func NewPhoneService(repo domain.PhoneRepository) PhoneService {
	return PhoneService{repo: repo}
}

func (service *PhoneService) RegisterPhone(ctx context.Context, phoneNumber domain.PhoneNumber, userId domain.AccountID) (*domain.Phone, error) {
	if existingPhone := service.repo.FindByPhoneNumber(ctx, phoneNumber); existingPhone != nil {
		return existingPhone, nil
	} else {
		newPhone := domain.NewPhone(phoneNumber, userId, "")
		return service.repo.Save(ctx, newPhone)
	}
}

func (service *PhoneService) UpdateFCMToken(ctx context.Context, id domain.PhoneId, token domain.FCMToken) (*domain.Phone, error) {
	if phone := service.repo.FindById(ctx, id); phone != nil {
		phone.UpdateFCMToken(token)
		return service.repo.Save(ctx, *phone)
	}
	return nil, errors.New("no phone with given id found")
}

func (service *PhoneService) RemovePhone(ctx context.Context, id domain.PhoneId) bool {
	return service.repo.Delete(ctx, id)
}

func (service *PhoneService) GetPhoneById(ctx context.Context, id domain.PhoneId) (*domain.Phone, error) {
	return service.repo.FindById(ctx, id), nil
}

func (service *PhoneService) GetPhoneByNumber(ctx context.Context, number domain.PhoneNumber) (*domain.Phone, error) {
	return service.repo.FindByPhoneNumber(ctx, number), nil
}
