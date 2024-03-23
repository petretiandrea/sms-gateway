package application

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sms-gateway/internal/domain"
)

type DeliveryNotificationService struct {
	repo            domain.DeliveryNotificationConfigRepository
	smsRepository   domain.Repository
	webhookNotifier domain.WebhookNotifier
	log             *zap.Logger
}

func NewDeliveryNotificationService(
	repo domain.DeliveryNotificationConfigRepository,
	smsRepository domain.Repository,
	webhookNotifier domain.WebhookNotifier,
) DeliveryNotificationService {
	return DeliveryNotificationService{
		repo:            repo,
		smsRepository:   smsRepository,
		webhookNotifier: webhookNotifier,
		log:             zap.L().Named("delivery_notification_service"),
	}
}

func (service *DeliveryNotificationService) UpdateDeliveryConfig(
	config domain.DeliveryNotificationConfig,
) (*domain.DeliveryNotificationConfig, error) {
	if _, err := service.repo.Save(config); err == nil {
		return &config, nil
	} else {
		return nil, err
	}
}

func (service *DeliveryNotificationService) DisableDeliveryNotification(
	id domain.AccountID,
) *domain.DeliveryNotificationConfig {
	if config := service.repo.FindById(id); config != nil {
		config.Enabled = false
		if _, err := service.repo.Save(*config); err == nil {
			return config
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (service *DeliveryNotificationService) NotifyDelivery(sms domain.Sms) error {
	if sms := service.smsRepository.FindById(sms.Id); sms == nil {
		return errors.New(fmt.Sprintf("Sms with id %s not found", sms.Id))
	} else {
		if sms.IsSent {
			if config := service.repo.FindById(sms.UserId); config != nil {
				if config.Enabled {
					if err := service.webhookNotifier.Notify(sms, config.WebhookURL); err != nil {
						return err
					}
					service.log.Info("Delivery notification sent", zap.String("smsId", string(sms.Id)))
				}
			}
		}
	}
	return nil
}
