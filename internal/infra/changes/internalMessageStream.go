package changes

import (
	"go.uber.org/zap"
	"sms-gateway/internal/domain"
	"time"
)

type InternalMessageStream struct {
	buffer   []domain.Sms
	stream   chan domain.Sms
	resumeAt *time.Time
}

func NewInternalMessageStream(bufferSize int, resumeAt *time.Time) domain.MessageStream {
	return &InternalMessageStream{
		buffer:   make([]domain.Sms, 0, bufferSize),
		stream:   make(chan domain.Sms, bufferSize),
		resumeAt: resumeAt,
	}
}

func (m *InternalMessageStream) Changes() <-chan domain.Sms {
	return m.stream
}

func (m *InternalMessageStream) Add(sms domain.Sms) {
	if m.resumeAt != nil && m.buffer != nil {
		m.buffer = append(m.buffer, sms)
	} else if m.stream != nil {
		select {
		case m.stream <- sms:
			zap.L().Info("Emit message change")
		default:
			zap.L().Error("Buffer full, dropping changes", zap.String("smsId", string(sms.Id)))
		}
	}
}

func (m *InternalMessageStream) Start() {
	// Implement resume stream logic here
	if m.resumeAt != nil {
		// handle resuming from the specific time
		// flush buffer to stream if needed
		for _, sms := range m.buffer {
			m.stream <- sms
		}
		m.buffer = nil
	}
}

func (m *InternalMessageStream) Close() {
	close(m.stream)
	m.buffer = nil // properly clear the buffer
}

var (
	_ domain.MessageStream = (*InternalMessageStream)(nil)
)
