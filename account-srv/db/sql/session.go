package sql

import (
	"time"

	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/account-srv/db/models/session"
	psession "github.com/xmc-dev/xmc/account-srv/proto/session"
)

func (s *SQL) CreateSession(sess *psession.Session) (*session.Session, error) {
	ss := session.Session{}
	ss.ClientID = sess.ClientId
	ss.CreatedAt = time.Unix(sess.CreatedAt, 0)
	ss.ExpiresAt = time.Unix(sess.ExpiresAt, 0)

	err := s.db.Create(&ss).Error

	return &ss, err
}

func (s *SQL) ReadSession(uuid uuid.UUID) (*session.Session, error) {
	ss := &session.Session{}

	err := s.db.First(ss, "uuid = ?", uuid).Error

	return ss, e(err)
}

func (s *SQL) DeleteSession(uuid uuid.UUID) error {
	ss := session.Session{}

	err := s.db.First(&ss, "uuid = ?", uuid).Error
	if err != nil {
		return e(err)
	}

	return e(s.db.Delete(&ss).Error)
}
