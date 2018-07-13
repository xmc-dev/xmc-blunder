package db

import (
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/queueitem"
)

type QueueItem interface {
	EnqueueJob(priority int, jobUUID uuid.UUID) error
	GetFirstJobInQueue() (*queueitem.QueueItem, error)
	DequeueJob() (*queueitem.QueueItem, error)
}

func EnqueueJob(priority int, jobUUID uuid.UUID) error {
	return db.EnqueueJob(priority, jobUUID)
}

func GetFirstJobInQueue() (*queueitem.QueueItem, error) {
	return db.GetFirstJobInQueue()
}

func DequeueJob() (*queueitem.QueueItem, error) {
	return db.DequeueJob()
}
