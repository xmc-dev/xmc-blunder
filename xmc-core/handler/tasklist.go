package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/xmc-core/db"
	mtasklist "github.com/xmc-dev/xmc/xmc-core/db/models/tasklist"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
	"github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"
	"github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
	"github.com/xmc-dev/xmc/xmc-core/proto/tsrange"
	"github.com/xmc-dev/xmc/xmc-core/util"
)

type TaskListService struct{}

func tasklistSName(method string) string {
	return fmt.Sprintf("%s.TaskListService.%s", "xmc.srv.core", method)
}

func taskListPage(title string, id uuid.UUID) string {
	return fmt.Sprintf(`# %s

{{macro "TaskList" "taskListId=%s"}}`, strings.Title(title), id.String())
}

func validateTimeRange(methodName string, timeRange *tsrange.TimestampRange) error {
	if timeRange == nil {
		return nil
	}
	if (timeRange.Begin == nil && timeRange.End != nil) || (timeRange.Begin != nil && timeRange.End == nil) {
		return errors.BadRequest(methodName, "start_time must have and end_time and vice versa")
	}
	start, _ := ptypes.Timestamp(timeRange.Begin)
	end, _ := ptypes.Timestamp(timeRange.End)
	if end.Before(start) {
		return errors.BadRequest(methodName, "end_time is before start_time")
	}
	return nil
}

func (*TaskListService) Create(ctx context.Context, req *tasklist.CreateRequest, rsp *tasklist.CreateResponse) error {
	methodName := tasklistSName("Create")
	switch {
	case req.TaskList == nil:
		return errors.BadRequest(methodName, "missing task list")
	case len(req.TaskList.Name) == 0:
		return errors.BadRequest(methodName, "invalid name")
	}
	if err := validateTimeRange(methodName, req.TaskList.TimeRange); err != nil {
		return err
	}

	req.TaskList.Id = ""
	req.TaskList.Name = strings.ToLower(req.TaskList.Name)
	if len(req.TaskList.Title) == 0 {
		req.TaskList.Title = strings.Title(req.TaskList.Name)
	}

	dd := db.DB.BeginGroup()
	id, err := dd.CreateTaskList(req.TaskList)
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Id = id.String()

	path := "/" + req.TaskList.Name
	pageID, err := util.CreatePage(dd, &page.CreateRequest{
		Page:     &page.Page{Path: path},
		Contents: taskListPage(req.TaskList.Title, id),
		Title:    req.TaskList.Title,
	})
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "page "+path+" already exists")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	err = dd.SetPageID(mtasklist.TaskList{}, id, pageID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*TaskListService) Read(ctx context.Context, req *tasklist.ReadRequest, rsp *tasklist.ReadResponse) error {
	methodName := tasklistSName("Read")
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	t, err := db.DB.ReadTaskList(id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task list not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.TaskList = t.ToProto()

	return nil
}

func (*TaskListService) Get(ctx context.Context, req *tasklist.GetRequest, rsp *tasklist.GetResponse) error {
	methodName := tasklistSName("Get")
	if len(req.Name) == 0 {
		return errors.BadRequest(methodName, "invalid name")
	}
	req.Name = strings.ToLower(req.Name)

	t, err := db.DB.GetTaskList(req.Name)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task list not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.TaskList = t.ToProto()
	return nil
}

func (*TaskListService) Update(ctx context.Context, req *tasklist.UpdateRequest, rsp *tasklist.UpdateResponse) error {
	methodName := tasklistSName("Update")

	_, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	err = db.DB.UpdateTaskList(req)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task list not found")
		} else if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*TaskListService) Delete(ctx context.Context, req *tasklist.DeleteRequest, rsp *tasklist.DeleteResponse) error {
	methodName := tasklistSName("Delete")
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	tl, err := dd.ReadTaskList(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task list not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	err = dd.DeleteTaskList(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "task list not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	err = util.DeletePage(dd, tl.PageID, true, log)
	if err != nil && err != db.ErrNotFound {
		dd.Rollback()
		return errors.InternalServerError(methodName, err.Error())
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*TaskListService) Search(ctx context.Context, req *tasklist.SearchRequest, rsp *tasklist.SearchResponse) error {
	methodName := tasklistSName("Search")

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	ts, total, err := db.DB.SearchTaskList(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	tls := []*tasklist.TaskList{}
	for _, t := range ts {
		tls = append(tls, t.ToProto())
	}

	rsp.TaskLists = tls
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(tls)),
		Total:   total,
	}
	return nil
}
