package sql

import (
	"github.com/jinzhu/gorm"
	"github.com/xmc-dev/xmc/dispatcher-srv/db"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/job"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/queueitem"
	"github.com/xmc-dev/xmc/dispatcher-srv/service"
	// db dialects
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type SQL struct {
	db *gorm.DB
}

func e(err error) error {
	if err == gorm.ErrRecordNotFound {
		return db.ErrNotFound
	}
	return err
}

func (s *SQL) lockTables(tx *gorm.DB, tables ...string) error {
	lck := ""
	for i, table := range tables {
		lck += table
		if i < len(tables)-1 {
			lck += ", "
		}
	}
	return tx.Exec("LOCK TABLE " + lck + " IN ACCESS EXCLUSIVE MODE;").Error
}

func (s *SQL) unlockTables(tx *gorm.DB) error {
	return nil
}

func (s *SQL) migrate() error {
	mq := s.db.AutoMigrate(job.Job{})
	if err := mq.Error; err != nil {
		return err
	}
	s.db.AutoMigrate(queueitem.QueueItem{}, queueitem.FinishedQueueItem{})
	if err := mq.Error; err != nil {
		return err
	}

	return nil
}

func (s *SQL) Init(dbType, dbURL string) error {
	var err error
	s.db, err = gorm.Open(dbType, dbURL)

	if err != nil {
		return err
	}

	if service.MainService.DBLog {
		s.db.LogMode(true)
	}

	return s.migrate()
}

func (s *SQL) Deinit() error {
	return s.db.Close()
}
