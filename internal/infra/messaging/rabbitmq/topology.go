package rabbitmq

const (
	ExchangeSMS = "sms"

	QueueSMSSendInternal      = "sms.send.internal"
	QueueSMSSendDeadLetter    = "sms.send.dead-letter"
	QueueSMSDeliveryConfirmed = "sms.delivery-confirmed"

	RoutingKeySMSSendInternalRequested = "sms.send.internal.requested"
	RoutingKeySMSSendDeadLetter        = "sms.send.dead-letter"
	RoutingKeySMSDeliveryConfirmed     = "sms.delivery.confirmed"
)
