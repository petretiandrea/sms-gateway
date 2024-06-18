package changes

import (
	"sms-gateway/internal/domain"
	"time"
)

type MessageChangeFeedProducer struct {
	openStreams []domain.MessageStream
}

func NewMessageChangeFeedProducer() *MessageChangeFeedProducer {
	return &MessageChangeFeedProducer{
		openStreams: make([]domain.MessageStream, 0), // Initialize as an empty slice
	}
}

func (m *MessageChangeFeedProducer) ResumeFrom(resumeAt *time.Time) domain.MessageStream {
	stream := NewInternalMessageStream(10, resumeAt)
	m.openStreams = append(m.openStreams, stream)
	go stream.(*InternalMessageStream).Start()
	return stream
}

func (m *MessageChangeFeedProducer) Add(sms domain.Sms) {
	for _, stream := range m.openStreams {
		go stream.(*InternalMessageStream).Add(sms)
	}
}

var (
	_ domain.MessageChangeFeedController = (*MessageChangeFeedProducer)(nil)
)
