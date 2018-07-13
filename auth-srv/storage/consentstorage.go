package storage

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
)

// ConsentRequest represents a user consent request for accessing oauth2 scopes
type ConsentRequest struct {
	Scope    string
	State    string
	ClientID string
}

// ConsentStorage handles storing consent requests in the database
type ConsentStorage struct {
	conn      redis.Conn
	keyPrefix string
}

// NewConsentStorage creates a new consent storage duh
func NewConsentStorage(keyPrefix string, conn redis.Conn) *ConsentStorage {
	c := &ConsentStorage{}
	c.conn = conn
	c.keyPrefix = keyPrefix

	return c
}

func (s *ConsentStorage) k(req *ConsentRequest) string {
	return fmt.Sprintf("%s:%s:%s", s.keyPrefix, req.ClientID, req.Scope)
}

// SaveRequest saves the request in the database.
// If an identical request is already present, it will be overwritten
func (s *ConsentStorage) SaveRequest(req *ConsentRequest) error {
	_, err := s.conn.Do("SET", s.k(req), req.State)

	return err
}

// ValidateRequest validates the consent request.
// Returns true if request is accepted, false if not.
func (s *ConsentStorage) ValidateRequest(req *ConsentRequest) bool {
	result, err := s.conn.Do("GET", s.k(req))
	bs, _ := redis.Bytes(result, err)
	logrus.WithFields(logrus.Fields{
		"req":    req,
		"key":    s.k(req),
		"result": string(bs),
	}).Debug("Validating request")
	if err != nil {
		if err != redis.ErrNil {
			logrus.WithField("err", err).Warn("Something happened when validating consent request")
		}
		return false
	}

	s.conn.Do("DEL", s.k(req))

	return string(bs) == req.State
}
