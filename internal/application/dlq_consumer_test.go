package application

import (
	"context"
	"testing"
	"time"

	"sms-gateway/internal/domain"
	"sms-gateway/internal/messages"

	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeDLQMessageRepository struct {
	message domain.DLQMessage
	called  bool
}

func (repo *fakeDLQMessageRepository) Save(ctx context.Context, message domain.DLQMessage) error {
	repo.message = message
	repo.called = true
	return nil
}

func TestDLQConsumerStoresDelivery(t *testing.T) {
	repo := &fakeDLQMessageRepository{}
	consumer := NewDLQConsumer(repo)
	occurredAt := time.Now()

	delivery := amqp.Delivery{
		MessageId:   "evt-1",
		Type:        string(messages.ChannelSMSSendInternal),
		Exchange:    "sms",
		RoutingKey:  "sms.send.dead-letter",
		ContentType: "application/json",
		Timestamp:   occurredAt,
		Headers:     amqp.Table{"affinity_key": "sms-1"},
		Body:        []byte(`{"messageId":"sms-1"}`),
	}

	if err := consumer.Process(context.Background(), delivery); err != nil {
		t.Fatalf("consumer.Process() error = %v", err)
	}

	if !repo.called {
		t.Fatal("expected repository to be called")
	}
	if repo.message.MessageID != "evt-1" {
		t.Fatalf("MessageID = %q, want %q", repo.message.MessageID, "evt-1")
	}
	if repo.message.Channel != string(messages.ChannelSMSSendInternal) {
		t.Fatalf("Channel = %q, want %q", repo.message.Channel, messages.ChannelSMSSendInternal)
	}
	if repo.message.RoutingKey != "sms.send.dead-letter" {
		t.Fatalf("RoutingKey = %q, want %q", repo.message.RoutingKey, "sms.send.dead-letter")
	}
	if string(repo.message.Payload) != `{"messageId":"sms-1"}` {
		t.Fatalf("Payload = %q", string(repo.message.Payload))
	}
	if repo.message.Metadata["affinity_key"] != "" {
		t.Fatalf("expected affinity key to be excluded from metadata")
	}
	if !repo.message.OccurredAt.Equal(occurredAt) {
		t.Fatalf("OccurredAt = %v, want %v", repo.message.OccurredAt, occurredAt)
	}
}
