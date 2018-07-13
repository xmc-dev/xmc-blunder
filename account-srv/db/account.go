package db

import (
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/account-srv/db/models/account"
	paccount "github.com/xmc-dev/xmc/account-srv/proto/account"
)

// Account represents a set of CRUD functions to interract with the Account database objects
type Account interface {
	CreateAccount(acc *paccount.Account) (uuid.UUID, error)
	ReadAccount(uuid uuid.UUID) (*account.Account, error)
	GetAccount(clientID string) (*account.Account, error)
	UpdateAccount(req *paccount.UpdateRequest) error
	DeleteAccount(uuid uuid.UUID) error
	SearchAccount(req *paccount.SearchRequest) ([]*account.Account, error)
	CreateRootAccount() (uuid.UUID, error)
}

// CreateAccount creates an account
func CreateAccount(acc *paccount.Account) (uuid.UUID, error) {
	return db.CreateAccount(acc)
}

// ReadAccount returns a single account that matches the query
func ReadAccount(uuid uuid.UUID) (*account.Account, error) {
	return db.ReadAccount(uuid)
}

func GetAccount(clientID string) (*account.Account, error) {
	return db.GetAccount(clientID)
}

// UpdateAccount updates account details
func UpdateAccount(req *paccount.UpdateRequest) error {
	return db.UpdateAccount(req)
}

// DeleteAccount deletes an account
func DeleteAccount(uuid uuid.UUID) error {
	return db.DeleteAccount(uuid)
}

// SearchAccount returns multiple candidates that match the query. Supports pagination
func SearchAccount(req *paccount.SearchRequest) ([]*account.Account, error) {
	return db.SearchAccount(req)
}

func CreateRootAccount() (uuid.UUID, error) {
	return db.CreateRootAccount()
}
