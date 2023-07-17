package sms

import (
	"sms-gateway/internal/account"
	"time"

	"github.com/google/uuid"
)

type MessageId string

type Message struct {
	Id             MessageId
	From           PhoneNumber
	To             string
	Content        string
	UserId         account.AccountID
	IsSent         bool
	SendAttempts   int
	CreatedAt      time.Time
	idempotencyKey string
}

type Repository interface {
	Save(message Message) (*Message, error)
	FindById(id MessageId) *Message
	FindExisting(idempotencyKey string) *Message
}

func CreateNewSMS(userId account.AccountID, from PhoneNumber, to PhoneNumber, content string, idempotencyKey string) Message {
	return Message{
		Id:             MessageId(uuid.NewString()),
		UserId:         userId,
		From:           from,
		To:             to.Number,
		Content:        content,
		IsSent:         false,
		SendAttempts:   0,
		CreatedAt:      time.Now(),
		idempotencyKey: idempotencyKey,
	}
}
