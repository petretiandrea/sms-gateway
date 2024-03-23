package events

import (
	"cloud.google.com/go/firestore"
	"go.uber.org/zap"
	"sms-gateway/internal/application"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/infra/repos"
	"time"
)

type DeliveryNotificationConsumer struct {
	eventListener  infra.FirestoreEventListener
	service        application.DeliveryNotificationService
	checkpointTime time.Time
	log            *zap.Logger
}

func NewDeliveryNotificationConsumer(eventListener infra.FirestoreEventListener, service application.DeliveryNotificationService) DeliveryNotificationConsumer {
	return DeliveryNotificationConsumer{
		eventListener:  eventListener,
		service:        service,
		checkpointTime: time.Now(),
		log:            zap.L().Named("delivery_notification_consumer"),
	}
}

func (consumer *DeliveryNotificationConsumer) Start() {
	consumer.log.Info("Resume checkpoint", zap.Time("checkpointTime", consumer.checkpointTime))
	go consumer.eventListener.ListenChanges(consumer.checkpointTime)

	for event := range consumer.eventListener.Changes() {
		if event.Kind == firestore.DocumentModified || event.Kind == firestore.DocumentAdded {
			var message repos.MessageFirestoreEntity
			if err := event.Doc.DataTo(&message); err == nil {
				if err := consumer.service.NotifyDelivery(*message.ToMessage(event.Doc.Ref.ID)); err != nil {
					consumer.log.Error("Failed to notify sms delivery", zap.Error(err))
				}
			} else {
				consumer.log.Error("Failed to decode sms structure from firebase changes", zap.Error(err), zap.String("documentId", event.Doc.Ref.ID))
			}
		}
		if event.Doc != nil {
			consumer.checkpointTime = event.Doc.UpdateTime
			consumer.log.Info("Saving checkpoint", zap.Time("checkpointTime", consumer.checkpointTime))
		}
	}
}

func (consumer *DeliveryNotificationConsumer) Stop() {
	consumer.eventListener.StopListenChanges()
}
