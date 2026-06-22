package messages

import (
	"encoding/json"
	"time"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

const (
	ChannelSMSAttemptInternal = outbox.Channel("internal.sms.attempt-requests")
)

type SMSAttemptRegistered struct {
	MessageID string `json:"messageId"`
}

func NewSMSAttemptRegistered(
	eventID string,
	occurredAt time.Time,
	messageID string,
) (outbox.Message, error) {
	payload, err := json.Marshal(SMSAttemptRegistered{
		MessageID: messageID,
	})
	if err != nil {
		return outbox.Message{}, err
	}

	return outbox.Message{
		ID:          eventID,
		Channel:     ChannelSMSAttemptInternal,
		AffinityKey: outbox.AffinityKey(messageID),
		Payload:     payload,
		OccurredAt:  occurredAt,
	}, nil
}

func UnmarshalSMSAttemptRegistered(data []byte) (SMSAttemptRegistered, error) {
	var message SMSAttemptRegistered
	err := json.Unmarshal(data, &message)
	return message, err
}
