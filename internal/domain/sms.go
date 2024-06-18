package domain

import (
	"github.com/pkg/errors"
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
	LastAttempt    Attempt
	CreatedAt      time.Time
	LastUpdateAt   time.Time
	IdempotencyKey string
	Configuration  WebhookConfiguration
	Metadata       map[string]string
}

type WebhookConfiguration struct {
	Url string
}

type QueryParams struct {
	From   string `form:"from" binding:"omitempty"`
	IsSent *bool  `form:"isSent" binding:"omitempty"`
}

type Repository interface {
	Save(message Sms) (*Sms, error)
	FindById(id SmsId) *Sms
	FindExisting(idempotencyKey string) *Sms
	Find(params QueryParams) ([]Sms, error)
}

func CreateNewSMS(userId AccountID, from PhoneNumber, to PhoneNumber, content string, idempotencyKey string, metadata map[string]string,
	configuration WebhookConfiguration) Sms {
	return Sms{
		Id:             SmsId(uuid.NewString()),
		UserId:         userId,
		From:           from,
		To:             to.Number,
		Content:        content,
		IsSent:         false,
		LastAttempt:    nil,
		CreatedAt:      time.Now(),
		IdempotencyKey: idempotencyKey,
		Metadata:       metadata,
		Configuration:  configuration,
	}
}

func (sms *Sms) RegisterAttempt(attempt Attempt) {
	var lastAttemptCount int32
	if sms.LastAttempt == nil {
		lastAttemptCount = 0
	} else {
		lastAttemptCount = sms.LastAttempt.AttemptNumber()
	}

	if lastAttemptCount < attempt.AttemptNumber() {
		if failure, ok := attempt.(FailedAttempt); ok {
			sms.LastAttempt = failure
			sms.IsSent = false
		} else if success, ok := attempt.(SuccessAttempt); ok {
			sms.LastAttempt = success
			sms.IsSent = true
		}
	}
}

var (
	ErrorNotMessageOwner = errors.New("Different message owner")
)
