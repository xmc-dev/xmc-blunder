package sql

import (
	"github.com/xmc-dev/xmc/account-srv/db"
	"github.com/xmc-dev/xmc/account-srv/db/models/role"
	prole "github.com/xmc-dev/xmc/account-srv/proto/role"
)

func (s *SQL) CreateRole(ro *prole.Role) error {
	r := role.Role{}
	r.ID = ro.Id
	r.Name = ro.Name
	r.Scope = ro.Scope

	err := s.db.Create(&r).Error
	if err != nil {
		return e(err)
	}

	return nil
}

func (s *SQL) ReadRole(id string) (*role.Role, error) {
	r := &role.Role{}

	err := s.db.First(r, "id = ?", id).Error

	return r, e(err)
}

func (s *SQL) UpdateRole(req *prole.UpdateRequest) error {
	u := make(map[string]interface{})

	if len(req.NewId) > 0 {
		err := s.db.Exec("UPDATE roles SET id = ? WHERE id = ?", req.NewId, req.Id).Error
		if err != nil {
			return e(err)
		}
		req.Id = req.NewId
	}
	if len(req.Name) > 0 {
		u["name"] = req.Name
	}
	if len(req.Scope) > 0 {
		u["scope"] = req.Scope
	}

	q := s.db.Model(&role.Role{}).Where("id = ?", req.Id).UpdateColumns(u)
	if q.Error != nil {
		return e(q.Error)
	}
	if q.RowsAffected == 0 {
		return db.ErrNotFound
	}

	return nil
}

func (s *SQL) DeleteRole(id string) error {
	q := s.db.Where("id = ?", id).Delete(&role.Role{})
	if q.Error != nil {
		return e(q.Error)
	}

	if q.RowsAffected == 0 {
		return db.ErrNotFound
	}

	return nil
}

func (s *SQL) SearchRole(req *prole.SearchRequest) ([]*role.Role, error) {
	result := []*role.Role{}
	query := s.db.Limit(req.Limit).Offset(req.Offset)
	if len(req.Id) > 0 {
		query = query.Where("id ~* ?", req.Id)
	}
	if len(req.Name) > 0 {
		query = query.Where("name ~* ?", req.Name)
	}
	if len(req.Scope) > 0 {
		query = query.Where("scope ~* ?", req.Scope)
	}

	err := query.Find(&result).Error

	if err != nil {
		return nil, e(err)
	}

	return result, nil

}
