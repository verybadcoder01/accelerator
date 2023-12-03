package session

import (
	"time"

	"github.com/google/uuid"
)

type JustToken struct { // I just sometimes need only this and I don't want to use map[string]string
	Token string `json:"token"`
}

type Session struct {
	Token   string    `json:"token"`
	ExpTime time.Time `json:"-"`
}

func NewSession(expTime time.Time) Session {
	return Session{Token: uuid.New().String(), ExpTime: expTime}
}

func (s *Session) IsExpired() bool {
	return s.ExpTime.Before(time.Now())
}
