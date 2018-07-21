package db

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/xmc-dev/xmc/xmc-core/db/models/problem"
	"github.com/xmc-dev/xmc/xmc-core/db/models/tasklist"
	ptasklist "github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
)

func (d *Datastore) CreateTaskList(tl *ptasklist.TaskList) (uuid.UUID, error) {
	t := tasklist.FromProto(tl)

	err := d.db.Create(t).Error
	return t.ID, e(err, "couldn't create task list")
}

func (d *Datastore) ReadTaskList(id uuid.UUID) (*tasklist.TaskList, error) {
	t := &tasklist.TaskList{}

	err := d.db.First(t, "id = ?", id).Error
	return t, e(err, "couldn't read task list")
}

func (d *Datastore) GetTaskList(name string) (*tasklist.TaskList, error) {
	t := &tasklist.TaskList{}

	err := d.db.First(t, "name = ?", name).Error
	return t, e(err, "coudln't read task list")
}

func (d *Datastore) TaskListExists(id uuid.UUID) (bool, error) {
	row := d.db.Raw("SELECT EXISTS(SELECT 1 FROM task_lists WHERE id = ?)", id).Row()
	result := false

	err := row.Scan(&result)

	return result, e(err, "couldn't check for task list existence")
}

func (d *Datastore) UpdateTaskList(tl *ptasklist.UpdateRequest) error {
	dd := d.begin()
	id, _ := uuid.Parse(tl.Id)

	t, err := dd.ReadTaskList(id)
	if err != nil {
		dd.Rollback()
		return err
	}

	if len(tl.Name) > 0 {
		t.Name = tl.Name
	}
	if len(tl.Description) > 0 {
		t.Description = tl.Description
	}
	if len(tl.Title) > 0 {
		t.Title = tl.Title
	}
	if tl.SetNullTime {
		t.StartTime = nil
		t.EndTime = nil
	}
	if tl.TimeRange != nil {
		st, _ := ptypes.Timestamp(tl.TimeRange.Begin)
		et, _ := ptypes.Timestamp(tl.TimeRange.End)
		t.StartTime = &st
		t.EndTime = &et
	}
	if tl.PublicSubmissions != nil {
		t.PublicSubmissions = tl.PublicSubmissions.Value
	}
	if tl.WithParticipations != nil {
		t.WithParticipations = tl.WithParticipations.Value
	}

	if err := dd.db.Save(t).Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't update task list")
	}

	return e(dd.Commit(), "couldn't update task list")
}

func (d *Datastore) DeleteTaskList(id uuid.UUID) error {
	err := d.db.Model(&problem.Task{}).Where("task_list_id = ?", id).Update("task_list_id", gorm.Expr("NULL")).Error
	if err != nil {
		return e(err, "couldn't orphan tasks of task list")
	}

	result := d.db.Where("id = ?", id).Delete(&tasklist.TaskList{})
	if result.Error != nil {
		return e(result.Error, "couldn't delete task list")
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *Datastore) SearchTaskList(req *ptasklist.SearchRequest) ([]*tasklist.TaskList, uint32, error) {
	dd := d.begin()
	ts := []*tasklist.TaskList{}
	// dis gon be good
	// orders the task lists by the following rules:
	//  - lists without a time range come first,
	//  - lists that are currently running come second
	//  - lists that are not running come third (TODO: upcoming lists should be before those who are finished)
	//  - all of them sorted by their name and their time ranges
	query := dd.db.Order(`
		name ASC,
		task_lists.start_time IS NULL DESC,
		(select now() <@ tstzrange(task_lists.start_time, task_lists.end_time)) DESC,
		task_lists.start_time ASC,
		task_lists.end_time ASC
	`)
	if len(req.Name) > 0 {
		query = query.Where("name ~* ?", req.Name)
	}
	if len(req.Description) > 0 {
		query = query.Where("description ~* ?", req.Description)
	}
	if req.IsPermanent != nil {
		if req.IsPermanent.Value {
			query = query.Where("start_time IS NULL")
		} else {
			query = query.Where("start_time IS NOT NULL")
		}
	}
	if req.TimeRange != nil {
		startTime, _ := ptypes.Timestamp(req.TimeRange.Begin)
		endTime, _ := ptypes.Timestamp(req.TimeRange.End)
		r := tsrange(startTime, endTime)
		query = query.Where("tstzrange(task_lists.start_time, task_lists.end_time) && ?", r)
	}
	if len(req.Title) > 0 {
		query = query.Where("title ~* ?", req.Title)
	}
	if req.PublicSubmissions != nil {
		query = query.Where("public_submissions = ?", req.PublicSubmissions.Value)
	}
	if req.WithParticipations != nil {
		query = query.Where("with_participations = ?", req.WithParticipations.Value)
	}
	var cnt uint32
	if err := query.Model(&ts).Count(&cnt).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search task lists")
	}
	query = query.Limit(req.Limit).Offset(req.Offset)
	if err := query.Find(&ts).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search task lists")
	}

	err := dd.Commit()

	return ts, cnt, e(err, "couldn't search task lists")
}

func (d *Datastore) CreateParticipation(taskListID, userID uuid.UUID) error {
	p := &tasklist.Participation{
		TaskListID: taskListID,
		UserID:     userID,
	}

	err := d.db.FirstOrCreate(p, p).Error
	return e(err, "couldn't create participation")
}

func (d *Datastore) CancelParticipation(taskListID, userID uuid.UUID) error {
	p := &tasklist.Participation{}
	dd := d.begin()
	err := dd.db.Where("task_list_id = ? AND user_id = ?", taskListID, userID).First(p).Error
	if err != nil {
		dd.Rollback()
		return e(err, "couldn't cancel participation")
	}
	if err := dd.db.Delete(p).Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't cancel participation")
	}

	return e(dd.Commit(), "couldn't cancel participation")
}

func (d *Datastore) GetTaskListParticipants(taskListID uuid.UUID) ([]*tasklist.Participation, error) {
	var ps []*tasklist.Participation
	err := d.db.Find(&ps, "task_list_id = ?", taskListID).Error
	if err != nil {
		return nil, e(err, "couldn't get task list participants")
	}

	return ps, nil
}
