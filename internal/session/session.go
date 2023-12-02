package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Token   string
	ExpTime time.Time
}

func NewSession(expTime time.Time) Session {
	return Session{Token: uuid.New().String(), ExpTime: expTime}
}

func (s *Session) IsExpired() bool {
	return s.ExpTime.Before(time.Now())
}
