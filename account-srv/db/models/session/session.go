package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/account-srv/proto/session"
)

// Session represents a user session in the database
type Session struct {
	UUID uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`

	ClientID string

	CreatedAt time.Time
	ExpiresAt time.Time
}

// ToProto represents the Session as a protobuf message
func (s *Session) ToProto() *session.Session {
	ps := &session.Session{
		Uuid:      s.UUID.String(),
		ClientId:  s.ClientID,
		CreatedAt: s.CreatedAt.Unix(),
		ExpiresAt: s.ExpiresAt.Unix(),
	}

	return ps
}
