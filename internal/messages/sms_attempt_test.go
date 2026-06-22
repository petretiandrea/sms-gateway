package messages

import (
	"testing"
	"time"
)

func TestNewSMSAttemptRegistered(t *testing.T) {
	message := NewSMSAttemptRegistered("evt-1", time.Unix(100, 0), "sms-1")

	if got := message.Metadata["type"]; got != MessageTypeSMSAttemptRegistered {
		t.Fatalf("message.Metadata[type] = %q, want %q", got, MessageTypeSMSAttemptRegistered)
	}

	payload, err := UnmarshalSMSAttemptRegistered(message.Payload)
	if err != nil {
		t.Fatalf("UnmarshalSMSAttemptRegistered() error = %v", err)
	}

	if payload.MessageID != "sms-1" {
		t.Fatalf("payload.MessageID = %q, want %q", payload.MessageID, "sms-1")
	}
}
