package db

import (
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/account-srv/db/models/session"
	psession "github.com/xmc-dev/xmc/account-srv/proto/session"
)

// Session represents a user session
type Session interface {
	CreateSession(sess *psession.Session) (*session.Session, error)
	ReadSession(uuid uuid.UUID) (*session.Session, error)
	DeleteSession(uuid uuid.UUID) error
}

func CreateSession(sess *psession.Session) (*session.Session, error) {
	return db.CreateSession(sess)
}

func ReadSession(uuid uuid.UUID) (*session.Session, error) {
	return db.ReadSession(uuid)
}

func DeleteSession(uuid uuid.UUID) error {
	return db.DeleteSession(uuid)
}
