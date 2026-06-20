package messages

import (
	"encoding/json"
	"time"
)

const (
	MessageTypeSMSSendRequested = "sms.send.requested"
)

type Envelope[T any] struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	OccurredAt  time.Time `json:"occurredAt"`
	AggregateID string    `json:"aggregateId"`
	Data        T         `json:"data"`
}

type SMSSendRequested struct {
	MessageID string `json:"messageId"`
	PhoneID   string `json:"phoneId"`
	AccountID string `json:"accountId"`
}

func NewSMSSendRequested(
	eventID string,
	occurredAt time.Time,
	messageID string,
	phoneID string,
	accountID string,
) Envelope[SMSSendRequested] {
	return Envelope[SMSSendRequested]{
		ID:          eventID,
		Type:        MessageTypeSMSSendRequested,
		OccurredAt:  occurredAt,
		AggregateID: messageID,
		Data: SMSSendRequested{
			MessageID: messageID,
			PhoneID:   phoneID,
			AccountID: accountID,
		},
	}
}

func MarshalEnvelope[T any](message Envelope[T]) ([]byte, error) {
	return json.Marshal(message)
}

func UnmarshalSMSSendRequested(data []byte) (Envelope[SMSSendRequested], error) {
	var message Envelope[SMSSendRequested]
	err := json.Unmarshal(data, &message)
	return message, err
}
