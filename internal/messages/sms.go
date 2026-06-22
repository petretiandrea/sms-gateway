package messages

import (
	"encoding/json"
	"time"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

const (
	MessageTypeSMSSendRequested = "sms.send.requested"
	ChannelSMSSendInternal      = outbox.Channel("sms.send.internal")
)

type SMSSendRequested struct {
	MessageID string `json:"messageId"`
}

func NewSMSSendRequested(
	eventID string,
	occurredAt time.Time,
	messageID string,
) outbox.Message {
	payload, err := json.Marshal(SMSSendRequested{
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
		Metadata:    outbox.Metadata{"type": MessageTypeSMSSendRequested},
		OccurredAt:  occurredAt,
	}
}

func UnmarshalSMSSendRequested(data []byte) (SMSSendRequested, error) {
	var message SMSSendRequested
	err := json.Unmarshal(data, &message)
	return message, err
}
