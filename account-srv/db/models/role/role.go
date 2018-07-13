package role

import prole "github.com/xmc-dev/xmc/account-srv/proto/role"

type Role struct {
	ID    string `gorm:"primary_key"`
	Name  string
	Scope string
}

func (r *Role) ToProto() *prole.Role {
	return &prole.Role{
		Id:    r.ID,
		Name:  r.Name,
		Scope: r.Scope,
	}
}
