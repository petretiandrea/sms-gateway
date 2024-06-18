package domain

import "time"

type MessageChangeFeedController interface {
	ResumeFrom(time *time.Time) MessageStream
	Add(sms Sms)
}

type MessageStream interface {
	Changes() <-chan Sms
	Close()
}
