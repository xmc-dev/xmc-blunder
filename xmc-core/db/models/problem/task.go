package problem

import (
	"github.com/google/uuid"
	ptask "github.com/xmc-dev/xmc/xmc-core/proto/task"
)

// Task is a problem that contestants / users have to accomplish.
// There can be many problems with the same statement and even with the same dataset.
// A problem can belong to a maximum of one contests, thus if there are contests with the same problem,
// each problem with have a distinct problem for it.
// A "problem" is a pair of a Page and a Dataset
type Task struct {
	ID         uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	DatasetID  uuid.UUID `gorm:"type:uuid"`
	PageID     uuid.UUID `gorm:"type:uuid;default:null"`
	TaskListID uuid.UUID `gorm:"type:uuid"`
	Name       string    `gorm:"unique_index"`
	// Internal description for administrators
	Description string
	InputFile   string
	OutputFile  string
	Title       string
}

func TaskFromProto(tk *ptask.Task) *Task {
	id, _ := uuid.Parse(tk.Id)
	dsID, _ := uuid.Parse(tk.DatasetId)
	pgID, _ := uuid.Parse(tk.PageId)
	tlID, _ := uuid.Parse(tk.TaskListId)
	t := &Task{
		ID:          id,
		DatasetID:   dsID,
		PageID:      pgID,
		TaskListID:  tlID,
		Name:        tk.Name,
		Description: tk.Description,
		InputFile:   tk.InputFile,
		OutputFile:  tk.OutputFile,
		Title:       tk.Title,
	}

	return t
}

func (t *Task) ToProto() *ptask.Task {
	ts := &ptask.Task{
		Id:          t.ID.String(),
		DatasetId:   t.DatasetID.String(),
		PageId:      t.PageID.String(),
		TaskListId:  t.TaskListID.String(),
		Name:        t.Name,
		Description: t.Description,
		InputFile:   t.InputFile,
		OutputFile:  t.OutputFile,
		Title:       t.Title,
	}

	return ts
}
