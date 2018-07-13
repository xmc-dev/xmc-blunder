package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/RangelReale/osin"
	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/auth-srv/account"
)

func encode(v interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decode(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

// XMCStorage implements the osin.Storage inteface
type XMCStorage struct {
	conn      redis.Conn
	keyPrefix string
}

func (s *XMCStorage) k(key, value string) string {
	return fmt.Sprintf("%s:%s:%s", s.keyPrefix, key, value)
}

func (s *XMCStorage) loadAccess(token, hash string) (*osin.AccessData, error) {
	logrus.Debug("loadAccess")
	raw, err := s.conn.Do("GET", s.k(hash, token))
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, osin.ErrNotFound
	}

	data := &osin.AccessData{
		Client: &Client{},
	}
	rawBytes := []byte(raw.([]uint8))
	err = decode(rawBytes, data)
	if err != nil {
		logrus.Debug(err)
		return nil, err
	}

	return data, nil
}

func NewXMCStorage(keyPrefix string, conn redis.Conn) *XMCStorage {
	c := &XMCStorage{}
	c.conn = conn
	c.keyPrefix = keyPrefix

	gob.Register(map[string]interface{}{})
	gob.Register(&Client{})
	gob.Register(osin.AuthorizeData{})
	gob.Register(osin.AccessData{})
	gob.Register(&UserData{})

	return c
}

// CloseDB closes the database connection
func (s *XMCStorage) CloseDB() {
	s.conn.Close()
}

func (s *XMCStorage) Clone() osin.Storage {
	logrus.Debug("Clone")
	return s
}

func (s *XMCStorage) Close() {
	logrus.Debug("Close")
}

func (s *XMCStorage) GetClient(id string) (osin.Client, error) {
	log := logrus.WithField("id", id)
	log.Debug("GetClient")
	if len(id) == 0 {
		return nil, osin.ErrNotFound
	}
	acc, err := account.GetClient(id)
	if err != nil {
		return nil, err
	}

	cl := Client{
		ID:          id,
		RedirectURI: acc.CallbackUrl,
		Original:    acc,
	}

	log.WithField("cl", cl).Debug("Got client")

	return &cl, nil
}

func (s *XMCStorage) SaveAuthorize(ad *osin.AuthorizeData) error {
	logrus.WithField("data", fmt.Sprintf("%+v", ad)).Debug("SaveAuthorize")
	b, err := encode(ad)
	if err != nil {
		return nil
	}
	_, err = s.conn.Do("SET", s.k("authorize", ad.Code), b)

	if err != nil {
		logrus.Debug(err)
	}

	return err
}

func (s *XMCStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	logrus.Debug("LoadAuthorize")
	raw, err := s.conn.Do("GET", s.k("authorize", code))
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, osin.ErrNotFound
	}

	data := &osin.AuthorizeData{
		Client: &Client{},
	}
	rawBytes := []byte(raw.([]uint8))
	err = decode(rawBytes, data)
	if err != nil {
		logrus.Debug(err)
		return nil, err
	}

	return data, nil
}

func (s *XMCStorage) RemoveAuthorize(code string) error {
	logrus.Debug("RemoveAuthorize")
	var err error
	if _, err = s.conn.Do("DEL", s.k("authorize", code)); err != nil {
		logrus.Debug(err)
	}

	return err
}

func (s *XMCStorage) SaveAccess(data *osin.AccessData) error {
	logrus.WithField("data", fmt.Sprintf("%+v", data)).Debug("SaveAccess")
	b, err := encode(data)
	if err != nil {
		logrus.Debug(err)
		return err
	}

	if _, err := s.conn.Do("SET", s.k("access", data.AccessToken), b); err != nil {
		logrus.Debug(err)
		return err
	}
	if data.RefreshToken != "" {
		if _, err := s.conn.Do("SET", s.k("refresh", data.RefreshToken), b); err != nil {
			logrus.Debug(err)
			return err
		}
	}

	return nil
}

func (s *XMCStorage) LoadAccess(token string) (*osin.AccessData, error) {
	return s.loadAccess(token, "access")
}

func (s *XMCStorage) RemoveAccess(token string) error {
	var err error
	if _, err = s.conn.Do("DEL", s.k("access", token)); err != nil {
		logrus.Debug(err)
	}

	return err
}

func (s *XMCStorage) LoadRefresh(token string) (*osin.AccessData, error) {
	logrus.Debug("LoadRefresh")
	return s.loadAccess(token, "refresh")
}

func (s *XMCStorage) RemoveRefresh(token string) error {
	logrus.Debug("RemoveRefresh")
	var err error
	if _, err = s.conn.Do("DEL", s.k("refresh", token)); err != nil {
		logrus.Debug(err)
	}

	return err
}
