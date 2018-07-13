package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/xmc-core/db"
	"github.com/xmc-dev/xmc/xmc-core/db/models/problem"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
	"github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"
	"github.com/xmc-dev/xmc/xmc-core/proto/task"
	"github.com/xmc-dev/xmc/xmc-core/util"
)

type TaskService struct{}

func taskPage(title string, id uuid.UUID) string {
	return fmt.Sprintf(`<TaskHeader taskId="%s" />

# %s

<TaskFooter taskId="%s" />`, id.String(), title, id.String())
}

func taskSName(method string) string {
	return fmt.Sprintf("%s.TaskService.%s", "xmc.srv.core", method)
}

func checkDatasetID(d *db.Datastore, methodName, datasetID string) error {
	did, err := uuid.Parse(datasetID)
	if err != nil {
		return errors.BadRequest(methodName, "invalid dataset_id")
	}
	ok, err := d.DatasetExists(did)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	if !ok {
		return errors.BadRequest(methodName, "dataset doesn't exist")
	}

	return nil
}

func checkTaskListID(d *db.Datastore, methodName, taskListID string) error {
	tlID, err := uuid.Parse(taskListID)
	if err != nil {
		return errors.BadRequest(methodName, "invalid task_list_id")
	}
	ok, err := d.TaskListExists(tlID)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	if !ok {
		return errors.BadRequest(methodName, "task list doesn't exist")
	}

	return nil
}

func (*TaskService) Create(ctx context.Context, req *task.CreateRequest, rsp *task.CreateResponse) error {
	methodName := taskSName("Create")
	switch {
	case req.Task == nil:
		return errors.BadRequest(methodName, "missing task")
	case len(req.Task.Name) == 0:
		return errors.BadRequest(methodName, "invalid name")
	case len(req.Task.InputFile) == 0:
		return errors.BadRequest(methodName, "invalid input_file")
	case len(req.Task.OutputFile) == 0:
		return errors.BadRequest(methodName, "invalid output_file")
	}

	dd := db.DB.BeginGroup()
	if err := checkDatasetID(dd, methodName, req.Task.DatasetId); err != nil {
		dd.Rollback()
		return err
	}
	if err := checkTaskListID(dd, methodName, req.Task.TaskListId); err != nil {
		dd.Rollback()
		return err
	}

	req.Task.Name = strings.ToLower(req.Task.Name)
	if len(req.Task.Title) == 0 {
		req.Task.Title = strings.Title(req.Task.Name)
	}

	id, err := dd.CreateTask(req.Task)
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Id = id.String()

	path := "/archive/" + req.Task.Name
	pageID, err := util.CreatePage(dd, &page.CreateRequest{
		Page:     &page.Page{Path: path},
		Contents: taskPage(req.Task.Title, id),
		Title:    req.Task.Title,
	})
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "page "+path+" already exists")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.SetPageID(problem.Task{}, id, pageID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*TaskService) Read(ctx context.Context, req *task.ReadRequest, rsp *task.ReadResponse) error {
	methodName := taskSName("Read")
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	t, err := db.DB.ReadTask(id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Task = t.ToProto()

	return nil
}

func (*TaskService) Get(ctx context.Context, req *task.GetRequest, rsp *task.GetResponse) error {
	methodName := taskSName("Get")
	if len(req.Name) == 0 {
		return errors.BadRequest(methodName, "invalid name")
	}

	req.Name = strings.ToLower(req.Name)
	t, err := db.DB.GetTask(req.Name)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Task = t.ToProto()

	return nil
}

func (*TaskService) Update(ctx context.Context, req *task.UpdateRequest, rsp *task.UpdateResponse) error {
	methodName := taskSName("Update")

	_, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	if len(req.DatasetId) > 0 {
		err = checkDatasetID(dd, methodName, req.DatasetId)
		if err != nil {
			dd.Rollback()
			return err
		}
	}
	if len(req.TaskListId) > 0 {
		err = checkTaskListID(dd, methodName, req.TaskListId)
		if err != nil {
			dd.Rollback()
			return err
		}
	}

	err = dd.UpdateTask(req)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task not found")
		} else if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*TaskService) Delete(ctx context.Context, req *task.DeleteRequest, rsp *task.DeleteResponse) error {
	methodName := taskSName("Delete")
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	t, err := dd.ReadTask(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	err = dd.DeleteTask(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	err = util.DeletePage(dd, t.PageID, true, log)
	// it shouldn't bail out if the page doesn't exist
	if err != nil && err != db.ErrNotFound {
		dd.Rollback()
		return errors.InternalServerError(methodName, err.Error())
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*TaskService) Search(ctx context.Context, req *task.SearchRequest, rsp *task.SearchResponse) error {
	methodName := taskSName("Search")

	if len(req.DatasetId) > 0 {
		_, err := uuid.Parse(req.DatasetId)
		if err != nil {
			return errors.BadRequest(methodName, "invalid dataset_id")
		}
	}

	if len(req.TaskListId) > 0 && req.TaskListId != "null" {
		_, err := uuid.Parse(req.TaskListId)
		if err != nil {
			return errors.BadRequest(methodName, "invalid task_list_id")
		}
	}

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	ts, total, err := db.DB.SearchTask(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	tks := []*task.Task{}
	for _, t := range ts {
		tks = append(tks, t.ToProto())
	}

	rsp.Tasks = tks
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(tks)),
		Total:   total,
	}
	return nil
}
