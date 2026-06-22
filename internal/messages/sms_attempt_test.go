package messages

import (
	"testing"
	"time"
)

func TestNewSMSAttemptRegistered(t *testing.T) {
	message, err := NewSMSAttemptRegistered("evt-1", time.Unix(100, 0), "sms-1")
	if err != nil {
		t.Fatalf("NewSMSAttemptRegistered() error = %v", err)
	}

	if got := message.Channel; got != ChannelSMSAttemptInternal {
		t.Fatalf("message.Channel = %q, want %q", got, ChannelSMSAttemptInternal)
	}

	payload, err := UnmarshalSMSAttemptRegistered(message.Payload)
	if err != nil {
		t.Fatalf("UnmarshalSMSAttemptRegistered() error = %v", err)
	}

	if payload.MessageID != "sms-1" {
		t.Fatalf("payload.MessageID = %q, want %q", payload.MessageID, "sms-1")
	}
}
