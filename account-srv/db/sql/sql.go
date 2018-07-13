package sql

import (
	"github.com/go-gormigrate/gormigrate"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	// db dialects
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/xmc-dev/xmc/account-srv/db"
	"github.com/xmc-dev/xmc/account-srv/db/models/account"
	"github.com/xmc-dev/xmc/account-srv/db/models/role"
	"github.com/xmc-dev/xmc/account-srv/db/models/session"
	"github.com/xmc-dev/xmc/account-srv/service"
)

type SQL struct {
	db  *gorm.DB
	srv *service.Service
}

func e(err error) error {
	if err == gorm.ErrRecordNotFound {
		return db.ErrNotFound
	}

	e, ok := err.(*pq.Error)
	if ok {
		switch e.Code.Name() {
		case "unique_violation":
			return db.ErrUniqueViolation
		case "foreign_key_violation":
			return db.ErrHasDependants(e.Table)
		}
	}
	return err
}

func (s *SQL) migrate() error {
	m := gormigrate.New(s.db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201803230015",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(account.Account{}, session.Session{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("accounts", "sessions").Error
			},
		},
		{
			ID: "201803230030",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(role.Role{}).Error; err != nil {
					return err
				}
				return tx.Model(&account.Account{}).AddForeignKey("role_id", "roles(id)", "RESTRICT", "CASCADE").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("roles").Error
			},
		},
		{
			ID: "201803240015",
			Migrate: func(tx *gorm.DB) error {
				return tx.Model(&account.Account{}).AddUniqueIndex("uix_accounts_client_id", "client_id").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&account.Account{}).RemoveIndex("uix_accounts_client_id").Error
			},
		},
		{
			ID: "201804050030",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE accounts ALTER COLUMN role_id SET DEFAULT NULL").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE accounts ALTER COLUMN role_id SET DEFAULT ?", uuid.Nil).Error
			},
		},
		{
			ID: "201807070015",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE accounts ALTER COLUMN owner_uuid TYPE uuid USING uuid").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&account.Account{}).ModifyColumn("owner_uuid", "text").Error
			},
		},
	})

	return m.Migrate()
}

func (s *SQL) Init(dbType, dbURL string) error {
	var err error
	s.db, err = gorm.Open(dbType, dbURL)

	if err != nil {
		return err
	}

	s.srv = service.MainService
	if s.srv.Debug {
		s.db.LogMode(true)
	}

	return s.migrate()
}

func (s *SQL) Deinit() error {
	return s.db.Close()
}
