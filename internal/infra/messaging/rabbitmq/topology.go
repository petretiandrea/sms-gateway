package rabbitmq

const (
	ExchangeSMS = "sms"

	QueueSMSSendInternal      = "sms.send.internal"
	QueueSMSSendDeadLetter    = "sms.send.dead-letter"
	QueueSMSDeliveryConfirmed = "sms.delivery-confirmed"

	RoutingKeyInternal             = "internal.sms"
	RoutingKeySMSSendDeadLetter    = "sms.send.dead-letter"
	RoutingKeySMSDeliveryConfirmed = "sms.delivery.confirmed"
)
