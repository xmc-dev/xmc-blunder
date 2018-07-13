package sql

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/queueitem"
)

func (s *SQL) EnqueueJob(priority int, jobUUID uuid.UUID) error {
	qi := queueitem.QueueItem{
		Priority: priority,
		JobUUID:  jobUUID,
	}
	return s.db.Create(&qi).Error
}

func (s *SQL) getFirstJobInQueue(tx *gorm.DB) (*queueitem.QueueItem, error) {
	qi := &queueitem.QueueItem{}
	err := tx.
		Joins("LEFT JOIN finished_queue_items on queue_items.id=finished_queue_items.id").
		Where("finished_queue_items.id IS NULL").
		Order("priority desc").
		First(qi).Error
	return qi, e(err)
}

func (s *SQL) GetFirstJobInQueue() (*queueitem.QueueItem, error) {
	tx := s.db.Begin()
	s.lockTables(tx, "queue_items", "finished_queue_items")
	qi, err := s.getFirstJobInQueue(tx)
	if err != nil {
		s.unlockTables(tx)
		tx.Rollback()
		return qi, err
	}
	s.unlockTables(tx)
	tx.Commit()
	return qi, nil
}

func (s *SQL) DequeueJob() (*queueitem.QueueItem, error) {
	tx := s.db.Begin()
	s.lockTables(tx, "queue_items", "finished_queue_items")
	qi, err := s.getFirstJobInQueue(tx)
	if err != nil {
		s.unlockTables(tx)
		tx.Rollback()
		return qi, err
	}
	sqi := &queueitem.FinishedQueueItem{
		QueueItem: *qi,
	}
	err = tx.Create(sqi).Error
	if err != nil {
		tx.Rollback()
		s.unlockTables(tx)
		return qi, e(err)
	}
	s.unlockTables(tx)
	tx.Commit()
	return qi, nil
}
