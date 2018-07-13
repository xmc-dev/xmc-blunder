package db

import (
	"github.com/xmc-dev/xmc/account-srv/db/models/role"
	prole "github.com/xmc-dev/xmc/account-srv/proto/role"
)

type Role interface {
	CreateRole(role *prole.Role) error
	ReadRole(id string) (*role.Role, error)
	UpdateRole(req *prole.UpdateRequest) error
	DeleteRole(id string) error
	SearchRole(req *prole.SearchRequest) ([]*role.Role, error)
}

func CreateRole(role *prole.Role) error {
	return db.CreateRole(role)
}

func ReadRole(id string) (*role.Role, error) {
	return db.ReadRole(id)
}

func UpdateRole(req *prole.UpdateRequest) error {
	return db.UpdateRole(req)
}

func DeleteRole(id string) error {
	return db.DeleteRole(id)
}

func SearchRole(req *prole.SearchRequest) ([]*role.Role, error) {
	return db.SearchRole(req)
}
