package api

import (
	"context"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"

	"github.com/google/uuid"
)

type PhoneApiController struct {
	Phone   application.PhoneService
	Account application.UserAccountService
}


// RegisterPhone implements StrictServerInterface.
func (c *PhoneApiController) RegisterPhone(ctx context.Context, request RegisterPhoneRequestObject) (RegisterPhoneResponseObject, error) {
	user := ctx.Value("user").(domain.UserAccount)
	if device, err := c.Phone.RegisterPhone(
		ctx,
		domain.PhoneNumber{Number: request.Body.Phone},
		user.Id,
	); err == nil {
		return RegisterPhone201JSONResponse(string(device.Id)), nil
	} else {
		return RegisterPhone400Response{}, nil
	}
}

// GetPhoneById implements StrictServerInterface.
func (c *PhoneApiController) GetPhoneById(ctx context.Context, request GetPhoneByIdRequestObject) (GetPhoneByIdResponseObject, error) {
	if device, err := c.Phone.GetPhoneById(ctx, domain.PhoneId(request.PhoneId)); device != nil {
		deviceToken := string(device.Token)
		return GetPhoneById200JSONResponse{
			Id:        uuid.MustParse(string(device.Id)),
			Phone:     device.Phone.Number,
			Account:   uuid.MustParse(string(device.UserId)),
			FcmToken:  &deviceToken,
			CreatedAt: device.CreatedAt,
			UpdatedAt: device.UpdatedAt,
		}, nil
	} else {
		if err == nil {
			return GetPhoneById404Response{}, nil
		} else {
			return GetPhoneById400Response{}, nil
		}
	}
}

// UpdateFcmToken implements StrictServerInterface.
func (c *PhoneApiController) UpdateFcmToken(
	ctx context.Context, 
	request UpdateFcmTokenRequestObject,
) (UpdateFcmTokenResponseObject, error) {
	if device, err := c.Phone.UpdateFCMToken(
		ctx,
		domain.PhoneId(request.PhoneId.String()),
		domain.FCMToken(*request.Body.Token),
	); err == nil {
		return UpdateFcmToken200JSONResponse(string(device.Id)), nil
	} else {
		return UpdateFcmToken400Response{}, nil
	}
}

