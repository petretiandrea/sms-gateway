package events

import (
	"go.uber.org/zap"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"time"
)

type MessageChangeFeedProcessor struct {
	changeFeedController domain.MessageChangeFeedController
	activeStream         domain.MessageStream
	service              application.DeliveryNotificationService
	checkpointTime       time.Time
	log                  *zap.Logger
}

func NewDeliveryNotificationConsumer(changeFeedController domain.MessageChangeFeedController, service application.DeliveryNotificationService) MessageChangeFeedProcessor {
	return MessageChangeFeedProcessor{
		changeFeedController: changeFeedController,
		service:              service,
		checkpointTime:       time.Now(),
		log:                  zap.L().Named("delivery_notification_consumer"),
	}
}

func (consumer MessageChangeFeedProcessor) Start() {
	consumer.log.Info("Resume checkpoint", zap.Time("checkpointTime", consumer.checkpointTime))
	consumer.activeStream = consumer.changeFeedController.ResumeFrom(&consumer.checkpointTime)
	defer consumer.activeStream.Close()
	for message := range consumer.activeStream.Changes() {
		if err := consumer.service.NotifyDelivery(message); err != nil {
			consumer.log.Error("Failed to notify sms delivery", zap.Error(err))
		}
		consumer.checkpointTime = time.Now()
		consumer.log.Info("Saving checkpoint", zap.Time("checkpointTime", consumer.checkpointTime))
	}
}

func (consumer MessageChangeFeedProcessor) Stop() {
	if consumer.activeStream != nil {
		consumer.activeStream.Close()
	}
}

//func (consumer *MessageChangeFeedProcessor) Start() {
//	consumer.log.Info("Resume checkpoint", zap.Time("checkpointTime", consumer.checkpointTime))
//	go consumer.eventListener.ListenChanges(consumer.checkpointTime)
//
//	//for event := range consumer.eventListener.Changes() {
//	//	if event.Kind == firestore.DocumentModified || event.Kind == firestore.DocumentAdded {
//	//		var message repos.MessageFirestoreEntity
//	//		if err := event.Doc.DataTo(&message); err == nil {
//	//			if err := consumer.service.NotifyDelivery(*message.ToMessage(event.Doc.Ref.ID)); err != nil {
//	//				consumer.log.Error("Failed to notify sms delivery", zap.Error(err))
//	//			}
//	//		} else {
//	//			consumer.log.Error("Failed to decode sms structure from firebase changes", zap.Error(err), zap.String("documentId", event.Doc.Ref.ID))
//	//		}
//	//	}
//	//	if event.Doc != nil {
//	//		consumer.checkpointTime = event.Doc.UpdateTime
//	//		consumer.log.Info("Saving checkpoint", zap.Time("checkpointTime", consumer.checkpointTime))
//	//	}
//	//}
//}

//func (consumer *MessageChangeFeedProcessor) Stop() {
//	consumer.eventListener.StopListenChanges()
//}
