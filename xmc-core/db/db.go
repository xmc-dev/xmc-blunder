package db

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"time"

	"github.com/jinzhu/gorm"
	"github.com/xmc-dev/xmc/xmc-core/service"
)

// Datastore manages the flow of the DB operations.
// It provides DB operations and transaction management through
// "transactions" and "groups".
//
// A "transaction" is a DB transaction used in the implementation of a DB operation.
// A "group" is a DB transaction for different, high-level DB operations declared
// in this package. A transaction can be in a group but not the other way around.
type Datastore struct {
	db            *gorm.DB
	inTransaction bool
	inGroup       bool
}

var DB *Datastore

var ErrNotFound = errors.New("db: not found")
var ErrUniqueViolation = errors.New("db: unique violation")

type ErrHasDependants string

func (hd ErrHasDependants) Error() string {
	return fmt.Sprintf("db: one or more %s depend on this object or have invalid dependencies", string(hd))
}

func e(err error, msg string) error {
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	e, ok := err.(*pq.Error)
	if ok {
		logrus.Debug("Error is of type " + e.Code.Name())
		switch e.Code.Name() {
		case "unique_violation":
			return ErrUniqueViolation
		case "foreign_key_violation":
			return ErrHasDependants(e.Table)
		}
	}

	if err != nil {
		err = errors.Wrap(err, msg)
	}
	return err
}

func tsrange(begin, end time.Time) string {
	var a, b string
	if !begin.IsZero() {
		a = begin.Format(time.RFC3339Nano)
	}
	if !end.IsZero() {
		b = end.Format(time.RFC3339Nano)
	}
	return "[" + a + "," + b + "]"
}

func rfk(tx *gorm.DB, table, field, dest string) error {
	key := tx.Dialect().BuildKeyName(table, field, dest, "foreign")
	return tx.Exec("ALTER TABLE test_cases DROP CONSTRAINT " + key).Error
}

func (d *Datastore) NotEmpty(table string) (bool, error) {
	v := false
	err := d.db.Raw("SELECT EXISTS (SELECT 1 from ?)", table).Row().Scan(&v)
	if err != nil {
		return false, e(err, "couldn't check if table "+table+"is not empty")
	}

	return v, nil
}

// begin marks the datastore session as being in an internal transaction, used
// for the implementation of a function.
func (d *Datastore) begin() *Datastore {
	if d.inTransaction {
		panic("Transaction already in progress")
	}
	dd := &Datastore{}
	*dd = *d
	if !d.inGroup {
		dd.db = d.db.Begin()
	}
	dd.inTransaction = true

	return dd
}

// BeginGroup marks the datastore session as being in a group.
func (d *Datastore) BeginGroup() *Datastore {
	if d.inTransaction || d.inGroup {
		panic("Group or transaction already in progress")
	}
	dd := &Datastore{}
	*dd = *d
	dd.db = d.db.Begin()
	dd.inGroup = true

	return dd
}

func (d *Datastore) Commit() error {
	if !d.inTransaction && !d.inGroup {
		panic("Cannot commit while not in a transaction or group")
	} else if d.inGroup && d.inTransaction {
		return nil // the operation will be completed outside the method
	}
	return d.db.Commit().Error
}

func (d *Datastore) Rollback() error {
	if !d.inTransaction && !d.inGroup {
		panic("Cannot rollback while not in a transaction or group")
	} else if d.inGroup && d.inTransaction {
		return nil // the operation will be completed outside the method
	}
	return d.db.Rollback().Error
}

func Init(s *service.Service) error {
	DB = &Datastore{}
	dbType := s.DBType
	dbURL := s.DBURL
	var err error
	DB.db, err = gorm.Open(dbType, dbURL)

	if err != nil {
		return errors.Wrap(err, "failed to open DB connection")
	}

	if service.MainService.Debug {
		DB.db.LogMode(true)
	}

	return DB.Migrate()
}

func Deinit() error {
	err := DB.db.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close DB connection")
	}

	return nil
}
