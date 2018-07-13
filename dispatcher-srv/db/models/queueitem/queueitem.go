package queueitem

import (
	"github.com/google/uuid"
)

// QueueItem is an item in the job queue. The QueueItem table is the actual queue.
type QueueItem struct {
	ID       uint `gorm:"primary_key"`
	Priority int
	JobUUID  uuid.UUID `gorm:"type:uuid"`
}

// FinishedQueueItem is a QueueItem that has been dequeued.
type FinishedQueueItem struct {
	QueueItem
}
