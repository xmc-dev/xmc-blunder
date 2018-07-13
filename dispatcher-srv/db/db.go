package db

import (
	"errors"

	"github.com/xmc-dev/xmc/dispatcher-srv/service"
)

// DB represents a database engine
type DB interface {
	Init(dbType, dbURL string) error
	Deinit() error
	Job
	QueueItem
}

var db DB
var srv *service.Service

var ErrNotFound = errors.New("not found")

func Register(d DB) {
	db = d
}

func Init(s *service.Service) error {
	srv = s
	return db.Init(srv.DBType, srv.DBURL)
}

func Deinit() error {
	return db.Deinit()
}
