package events

import (
	"cloud.google.com/go/firestore"
	"go.uber.org/zap"
	"sms-gateway/internal/application"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/infra/repos"
)

type DeliveryNotificationConsumer struct {
	eventListener infra.FirestoreEventListener
	service       application.DeliveryNotificationService
}

func NewDeliveryNotificationConsumer(eventListener infra.FirestoreEventListener, service application.DeliveryNotificationService) DeliveryNotificationConsumer {
	return DeliveryNotificationConsumer{
		eventListener: eventListener,
		service:       service,
	}
}

func (consumer *DeliveryNotificationConsumer) Start() {
	go consumer.eventListener.ListenChanges()

	for event := range consumer.eventListener.Changes() {
		if event.Kind == firestore.DocumentModified || event.Kind == firestore.DocumentAdded {
			var message repos.MessageFirestoreEntity
			if err := event.Doc.DataTo(&message); err == nil {
				if err := consumer.service.NotifyDelivery(*message.ToMessage(event.Doc.Ref.ID)); err != nil {
					zap.L().Error("Failed to notify sms delivery", zap.Error(err))
				} else {
					zap.L().Info("Notify sms delivery successfully")
				}
			} else {
				zap.L().Error("Failed to decode sms structure from firebase changes", zap.Error(err))
			}
		}
	}
}

func (consumer *DeliveryNotificationConsumer) Stop() {
	consumer.eventListener.StopListenChanges()
}
