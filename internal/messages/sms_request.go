package messages

import (
	"encoding/json"
	"time"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

const (
	ChannelSMSSendInternal = outbox.Channel("internal.sms.send-requests")
)

type SMSSendRequested struct {
	MessageID string `json:"messageId"`
}

func NewSMSSendRequested(
	eventID string,
	occurredAt time.Time,
	messageID string,
) (outbox.Message, error) {
	payload, err := json.Marshal(SMSSendRequested{
		MessageID: messageID,
	})
	if err != nil {
		return outbox.Message{}, err
	}

	return outbox.Message{
		ID:          eventID,
		Channel:     ChannelSMSSendInternal,
		AffinityKey: outbox.AffinityKey(messageID),
		Payload:     payload,
		OccurredAt:  occurredAt,
	}, nil
}

func UnmarshalSMSSendRequested(data []byte) (SMSSendRequested, error) {
	var message SMSSendRequested
	err := json.Unmarshal(data, &message)
	return message, err
}
