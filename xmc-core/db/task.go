package db

import (
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/xmc-core/db/models/problem"
	ptask "github.com/xmc-dev/xmc/xmc-core/proto/task"
)

func (d *Datastore) CreateTask(tk *ptask.Task) (uuid.UUID, error) {
	t := problem.TaskFromProto(tk)

	err := d.db.Create(t).Error
	return t.ID, e(err, "couldn't create task")
}

func (d *Datastore) ReadTask(id uuid.UUID) (*problem.Task, error) {
	t := &problem.Task{}

	err := d.db.First(t, "id = ?", id).Error
	return t, e(err, "couldn't read task")
}

func (d *Datastore) GetTask(name string) (*problem.Task, error) {
	t := &problem.Task{}

	err := d.db.First(t, "name = ?", name).Error
	return t, e(err, "couldn't get task by name")
}

func (d *Datastore) UpdateTask(tk *ptask.UpdateRequest) error {
	dd := d.begin()
	id, _ := uuid.Parse(tk.Id)
	datasetID, _ := uuid.Parse(tk.DatasetId)
	taskListID, _ := uuid.Parse(tk.TaskListId)

	t, err := dd.ReadTask(id)
	if err != nil {
		dd.Rollback()
		return err
	}

	if len(tk.DatasetId) > 0 {
		t.DatasetID = datasetID
	}
	if len(tk.Name) > 0 {
		t.Name = tk.Name
	}
	if len(tk.Description) > 0 {
		t.Description = tk.Description
	}
	if len(tk.InputFile) > 0 {
		t.InputFile = tk.InputFile
	}
	if len(tk.OutputFile) > 0 {
		t.OutputFile = tk.OutputFile
	}
	if len(tk.Title) > 0 {
		t.Title = tk.Title
	}
	if len(tk.TaskListId) > 0 {
		t.TaskListID = taskListID
	}

	if err := dd.db.Save(t).Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't update task")
	}

	return e(dd.Commit(), "couldn't update task")
}

func (d *Datastore) DeleteTask(id uuid.UUID) error {
	result := d.db.Where("id = ?", id).Delete(&problem.Task{})
	if result.Error != nil {
		return e(result.Error, "couldn't delete task")
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *Datastore) SearchTask(req *ptask.SearchRequest) ([]*problem.Task, uint32, error) {
	dd := d.begin()
	ts := []*problem.Task{}
	query := dd.db
	if len(req.DatasetId) > 0 {
		datasetID, _ := uuid.Parse(req.DatasetId)
		query = query.Where("dataset_id = ?", datasetID)
	}
	if len(req.Name) > 0 {
		query = query.Where("name ~* ?", req.Name)
	}
	if len(req.Description) > 0 {
		query = query.Where("description ~* ?", req.Description)
	}
	if len(req.Title) > 0 {
		query = query.Where("title ~* ?", req.Title)
	}
	if len(req.TaskListId) > 0 {
		if req.TaskListId != "null" {
			tlID, _ := uuid.Parse(req.TaskListId)
			query = query.Where("task_list_id = ?", tlID)
		} else {
			query = query.Where("task_list_id IS NULL")
		}
	}
	var cnt uint32
	if err := query.Model(&ts).Count(&cnt).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search tasks")
	}
	query = query.Limit(req.Limit).Offset(req.Offset)
	if err := query.Find(&ts).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search tasks")
	}

	err := dd.Commit()

	return ts, cnt, e(err, "couldn't search tasks")
}
