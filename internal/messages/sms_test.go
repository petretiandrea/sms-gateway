package messages

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSMSSendRequestedJSONContract(t *testing.T) {
	occurredAt := time.Date(2026, 6, 20, 9, 0, 0, 0, time.UTC)
	message := NewSMSSendRequested(
		"evt_123",
		occurredAt,
		"msg_123",
		"phone_123",
		"account_123",
	)

	data, err := MarshalEnvelope(message)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw json: %v", err)
	}

	if raw["type"] != MessageTypeSMSSendRequested {
		t.Fatalf("unexpected type: %v", raw["type"])
	}
	if raw["aggregateId"] != "msg_123" {
		t.Fatalf("unexpected aggregateId: %v", raw["aggregateId"])
	}

	decoded, err := UnmarshalSMSSendRequested(data)
	if err != nil {
		t.Fatalf("unmarshal send requested: %v", err)
	}

	if decoded.Data.MessageID != "msg_123" {
		t.Fatalf("unexpected message id: %s", decoded.Data.MessageID)
	}
	if decoded.Data.PhoneID != "phone_123" {
		t.Fatalf("unexpected phone id: %s", decoded.Data.PhoneID)
	}
	if decoded.Data.AccountID != "account_123" {
		t.Fatalf("unexpected account id: %s", decoded.Data.AccountID)
	}
	if !decoded.OccurredAt.Equal(occurredAt) {
		t.Fatalf("unexpected occurredAt: %s", decoded.OccurredAt)
	}
}
