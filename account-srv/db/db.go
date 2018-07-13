package db

import (
	"errors"
	"fmt"
)

// DB represents a database engine
type DB interface {
	Init(dbType, dbURL string) error
	Deinit() error
	Account
	Session
	Role
}

var db DB

// ErrNotFound is returned when a query doesn't find anything
var ErrNotFound = errors.New("db: not found")
var ErrUniqueViolation = errors.New("db: unique violation")

type ErrHasDependants string

func (hd ErrHasDependants) Error() string {
	return fmt.Sprintf("db: one or more %s depend on this object or have invalid dependencies", string(hd))
}

// Register a database engine
func Register(d DB) {
	db = d
}

// Init initializes the database engine
func Init(dbType, dbURL string) error {
	return db.Init(dbType, dbURL)
}

// Deinit closes de database connection
func Deinit() error {
	return db.Deinit()
}
