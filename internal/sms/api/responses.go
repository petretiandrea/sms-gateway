package api

import (
	"time"
)

type SmsEntityResponse struct {
	Id           string    `json:"id" binding:"required"`
	Content      string    `json:"content" binding:"required"`
	From         string    `json:"from" binding:"required"`
	To           string    `json:"to" binding:"required"`
	UserId       string    `json:"owner" binding:"required"`
	IsSent       bool      `json:"isSent" binding:"required"`
	SendAttempts int       `json:"sendAttempts" binding:"required"`
	CreatedAt    time.Time `json:"createdAt" binding:"required"`
}
