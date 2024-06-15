package domain

type Attempt interface {
	AttemptNumber() int32
}

type SuccessAttempt struct {
	AttemptCount int32
	PhoneId      PhoneId
}

type FailedAttempt struct {
	AttemptCount int32
	PhoneId      PhoneId
	Reason       string
}

func (s SuccessAttempt) AttemptNumber() int32 {
	return s.AttemptCount
}

func (f FailedAttempt) AttemptNumber() int32 {
	return f.AttemptCount
}

var (
	_ Attempt = (*FailedAttempt)(nil)
	_ Attempt = (*SuccessAttempt)(nil)
)
