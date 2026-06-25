package domain

import (
	"context"
	"time"
)

type DLQMessage struct {
	MessageID   string
	Channel     string
	Exchange    string
	RoutingKey  string
	ContentType string
	Payload     []byte
	Metadata    map[string]string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

type DLQMessageRepository interface {
	Save(ctx context.Context, message DLQMessage) error
}
