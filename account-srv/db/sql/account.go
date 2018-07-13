package sql

import (
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/account-srv/db/models/account"
	"github.com/xmc-dev/xmc/account-srv/db/models/role"
	paccount "github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/account-srv/util"
)

func (s *SQL) CreateAccount(acc *paccount.Account) (uuid.UUID, error) {
	a := account.Account{}
	// UUID is given by the DB
	a.Type = account.Type(acc.Type)
	a.ClientID = acc.ClientId
	a.ClientSecret = acc.ClientSecret
	a.Name = acc.Name
	a.OwnerUUID, _ = uuid.Parse(acc.OwnerUuid)
	a.CallbackURL = acc.CallbackUrl
	a.IsFirstParty = false
	a.Scope = acc.Scope
	a.IsPublic = acc.IsPublic
	a.RoleID = acc.RoleId

	err := s.db.Create(&a).Error
	if err != nil {
		return uuid.Nil, e(err)
	}

	return a.UUID, nil
}

func (s *SQL) ReadAccount(uuid uuid.UUID) (*account.Account, error) {
	a := &account.Account{}

	err := s.db.First(a, "uuid = ?", uuid).Error

	return a, e(err)
}

func (s *SQL) GetAccount(clientID string) (*account.Account, error) {
	a := &account.Account{}

	err := s.db.First(a, "client_id = ?", clientID).Error

	return a, e(err)
}

func (s *SQL) UpdateAccount(req *paccount.UpdateRequest) error {
	a := account.Account{}
	err := s.db.First(&a, "uuid = ?", req.Uuid).Error

	if err != nil {
		return e(err)
	}

	if req.ClientSecret != "" {
		a.ClientSecret = req.ClientSecret
	}

	if req.CallbackUrl != "" {
		a.CallbackURL = req.CallbackUrl
	}

	if req.Name != "" {
		a.Name = req.Name
	}

	if req.Scope != "" {
		a.Scope = req.Scope
	}

	if req.RoleId != "" {
		a.RoleID = req.RoleId
	}

	return e(s.db.Save(&a).Error)
}

func (s *SQL) DeleteAccount(uuid uuid.UUID) error {
	a := account.Account{}

	err := s.db.First(&a, "uuid = ?", uuid).Error
	if err != nil {
		return e(err)
	}

	return e(s.db.Delete(&a).Error)
}

func (s *SQL) SearchAccount(req *paccount.SearchRequest) ([]*account.Account, error) {
	where := make(map[string]interface{})
	if len(req.ClientId) > 0 {
		where["client_id"] = req.ClientId
	}
	if req.Type != nil {
		where["type"] = req.Type.GetValue()
	}
	if req.IsPublic != nil {
		where["is_public"] = req.IsPublic.GetValue()
	}
	if len(req.OwnerUuid) > 0 {
		where["owner_uuid"] = req.OwnerUuid
	}
	if len(req.CallbackUrl) > 0 {
		where["callback_url"] = req.CallbackUrl
	}

	result := []*account.Account{}
	query := s.db.Where(where)
	if req.Limit > 0 {
		query = query.Limit(int(req.Limit))
	}
	if req.Offset > 0 {
		query = query.Offset(int(req.Offset))
	}
	if len(req.Name) > 0 {
		query = query.Where("name ~* ?", req.Name)
	}
	if len(req.RoleId) > 0 {
		query = query.Where("role_id = ?", req.RoleId)
	}
	err := query.Find(&result).Error

	if err != nil {
		return nil, e(err)
	}

	return result, nil
}

func (s *SQL) CreateRootAccount() (uuid.UUID, error) {
	tx := s.db.Begin()
	// create admin role
	r := role.Role{
		ID:    "admin",
		Scope: "*",
	}
	if err := tx.FirstOrCreate(&r, r).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}
	acc := account.Account{
		RoleID:   "admin",
		Type:     account.USER,
		ClientID: "root",
		Name:     "root",
	}
	if err := tx.FirstOrCreate(&acc, acc).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}
	if len(r.Name) == 0 {
		r.Name = "Administrator"
		if err := tx.Save(&r).Error; err != nil {
			tx.Rollback()
			return uuid.Nil, err
		}
	}
	if len(acc.ClientSecret) == 0 {
		hash, err := util.HashSecret("root")
		if err != nil {
			tx.Rollback()
			return uuid.Nil, err
		}
		acc.ClientSecret = hash
		if err := tx.Model(&acc).Update("client_secret", hash).Error; err != nil {
			tx.Rollback()
			return uuid.Nil, err
		}
	}
	botAcc := account.Account{
		Type: account.SERVICE,

		// Highly illegal. Some are born more equal than the others :)
		ClientID: "rootbot",

		Name:         "Terminator",
		OwnerUUID:    acc.UUID,
		CallbackURL:  "http://localhost/",
		IsFirstParty: true,
		IsPublic:     false,
		Scope:        "*",
	}
	if err := tx.FirstOrCreate(&botAcc, botAcc).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}
	if len(botAcc.ClientSecret) == 0 {
		hash, err := util.HashSecret("rootbot")
		if err != nil {
			tx.Rollback()
			return uuid.Nil, err
		}
		botAcc.ClientSecret = hash
		if err := tx.Model(&botAcc).Update("client_secret", hash).Error; err != nil {
			tx.Rollback()
			return uuid.Nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return uuid.Nil, err
	}

	return acc.UUID, nil
}
