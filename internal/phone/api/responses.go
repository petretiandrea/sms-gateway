package api

import "time"

type PhoneResponse struct {
	Id        string    `json:"id" binding:"required"`
	Phone     string    `json:"phone" binding:"required"`
	Account   string    `json:"account" binding:"required"`
	FCMToken  string    `json:"fcmToken" binding:"required"`
	CreatedAt time.Time `json:"createdAt" binding:"required"`
	UpdatedAt time.Time `json:"updatedAt" binding:"required"`
}
