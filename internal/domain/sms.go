package domain

import (
	"time"

	"github.com/google/uuid"
)

type SmsId string

type Sms struct {
	Id             SmsId
	From           PhoneNumber
	To             string
	Content        string
	UserId         AccountID
	IsSent         bool
	SendAttempts   int
	CreatedAt      time.Time
	IdempotencyKey string
}

type Repository interface {
	Save(message Sms) (*Sms, error)
	FindById(id SmsId) *Sms
	FindExisting(idempotencyKey string) *Sms
}

func CreateNewSMS(userId AccountID, from PhoneNumber, to PhoneNumber, content string, idempotencyKey string) Sms {
	return Sms{
		Id:             SmsId(uuid.NewString()),
		UserId:         userId,
		From:           from,
		To:             to.Number,
		Content:        content,
		IsSent:         false,
		SendAttempts:   0,
		CreatedAt:      time.Now(),
		IdempotencyKey: idempotencyKey,
	}
}
