package tasklist

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	ptasklist "github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
	ptsrange "github.com/xmc-dev/xmc/xmc-core/proto/tsrange"
)

// TaskList is a list of tasks with a wiki page. Users can solve the tasks
// and get a rank.
type TaskList struct {
	ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	PageID      uuid.UUID `gorm:"type:uuid"`
	StartTime   *time.Time
	EndTime     *time.Time
	Name        string
	Description string
	Title       string
}

func FromProto(tl *ptasklist.TaskList) *TaskList {
	id, _ := uuid.Parse(tl.Id)
	pgID, _ := uuid.Parse(tl.PageId)
	var startTime, endTime *time.Time
	if tl.TimeRange != nil {
		st, _ := ptypes.Timestamp(tl.TimeRange.Begin)
		et, _ := ptypes.Timestamp(tl.TimeRange.End)
		startTime = &st
		endTime = &et
	}
	t := &TaskList{
		ID:          id,
		PageID:      pgID,
		StartTime:   startTime,
		EndTime:     endTime,
		Name:        tl.Name,
		Description: tl.Description,
		Title:       tl.Title,
	}

	return t
}

func (t *TaskList) ToProto() *ptasklist.TaskList {
	tl := &ptasklist.TaskList{
		Id:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		PageId:      t.PageID.String(),
		Title:       t.Title,
	}
	tl.TimeRange = &ptsrange.TimestampRange{}
	if t.StartTime != nil && t.EndTime != nil {
		tl.TimeRange.Begin, _ = ptypes.TimestampProto(*t.StartTime)
		tl.TimeRange.End, _ = ptypes.TimestampProto(*t.EndTime)
	}

	return tl
}
