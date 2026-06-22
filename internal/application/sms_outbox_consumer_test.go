package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"sms-gateway/internal/messages"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeSMSSendRequestedHandler struct {
	called bool
}

func (handler *fakeSMSSendRequestedHandler) Handle(context.Context, outbox.Message) error {
	handler.called = true
	return nil
}

type fakeSMSAttemptRegisteredHandler struct {
	called bool
}

func (handler *fakeSMSAttemptRegisteredHandler) Handle(context.Context, outbox.Message) error {
	handler.called = true
	return nil
}

func TestSMSOutboxConsumerRoutesSMSSendRequested(t *testing.T) {
	sendHandler := &fakeSMSSendRequestedHandler{}
	consumer := NewSMSOutboxConsumer(sendHandler, &fakeSMSAttemptRegisteredHandler{})

	delivery := amqp.Delivery{
		MessageId: "evt-1",
		Type:      string(messages.ChannelSMSSendInternal),
		Timestamp: time.Now(),
		Headers:   amqp.Table{"type": messages.MessageTypeSMSSendRequested},
		Body:      []byte(`{"messageId":"sms-1"}`),
	}

	if err := consumer.Process(context.Background(), delivery); err != nil {
		t.Fatalf("consumer.Process() error = %v", err)
	}
	if !sendHandler.called {
		t.Fatal("expected send handler to be called")
	}
}

func TestSMSOutboxConsumerRoutesSMSAttemptRegistered(t *testing.T) {
	attemptHandler := &fakeSMSAttemptRegisteredHandler{}
	consumer := NewSMSOutboxConsumer(&fakeSMSSendRequestedHandler{}, attemptHandler)

	delivery := amqp.Delivery{
		MessageId: "evt-2",
		Type:      string(messages.ChannelSMSSendInternal),
		Timestamp: time.Now(),
		Headers:   amqp.Table{"type": messages.MessageTypeSMSAttemptRegistered},
		Body:      []byte(`{"messageId":"sms-1"}`),
	}

	if err := consumer.Process(context.Background(), delivery); err != nil {
		t.Fatalf("consumer.Process() error = %v", err)
	}
	if !attemptHandler.called {
		t.Fatal("expected attempt handler to be called")
	}
}

func TestSMSOutboxConsumerRejectsUnknownMessageType(t *testing.T) {
	consumer := NewSMSOutboxConsumer(&fakeSMSSendRequestedHandler{}, &fakeSMSAttemptRegisteredHandler{})
	delivery := amqp.Delivery{
		MessageId: "evt-1",
		Type:      string(messages.ChannelSMSSendInternal),
		Timestamp: time.Now(),
		Headers:   amqp.Table{"type": "unknown.event"},
		Body:      []byte(`{"messageId":"sms-1"}`),
	}

	err := consumer.Process(context.Background(), delivery)
	if err == nil {
		t.Fatal("expected error")
	}

	var retryable interface{ Retryable() bool }
	if !errors.As(err, &retryable) || retryable.Retryable() {
		t.Fatalf("expected non-retryable error, got %v", err)
	}
}
