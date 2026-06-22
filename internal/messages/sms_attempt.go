package messages

import (
	"encoding/json"
	"time"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

const (
	MessageTypeSMSAttemptRegistered = "sms.attempt.registered"
)

type SMSAttemptRegistered struct {
	MessageID string `json:"messageId"`
}

func NewSMSAttemptRegistered(
	eventID string,
	occurredAt time.Time,
	messageID string,
) outbox.Message {
	payload, err := json.Marshal(SMSAttemptRegistered{
		MessageID: messageID,
	})
	if err != nil {
		panic(err) // TODO: avoid that
	}

	return outbox.Message{
		ID:          eventID,
		Channel:     ChannelSMSSendInternal,
		AffinityKey: outbox.AffinityKey(messageID),
		Payload:     payload,
		Metadata:    outbox.Metadata{"type": MessageTypeSMSAttemptRegistered},
		OccurredAt:  occurredAt,
	}
}

func UnmarshalSMSAttemptRegistered(data []byte) (SMSAttemptRegistered, error) {
	var message SMSAttemptRegistered
	err := json.Unmarshal(data, &message)
	return message, err
}
