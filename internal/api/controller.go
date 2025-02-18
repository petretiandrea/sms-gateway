package api

type Controller struct {
	AttemptController
	SmsApiController
	PhoneApiController
	UserAccountController
	DeliveryNotificationController
}

var _ StrictServerInterface = (*Controller)(nil)
