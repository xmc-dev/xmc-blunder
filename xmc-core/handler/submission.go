package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/dispatcher-srv/proto/job"
	"github.com/xmc-dev/xmc/xmc-core/common"
	"github.com/xmc-dev/xmc/xmc-core/db"
	"github.com/xmc-dev/xmc/xmc-core/db/models/tasklist"
	"github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	"github.com/xmc-dev/xmc/xmc-core/proto/result"
	"github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"
	"github.com/xmc-dev/xmc/xmc-core/proto/submission"
	"github.com/xmc-dev/xmc/xmc-core/util"
)

type SubmissionService struct{}

func timeInRange(startTime, now, endTime time.Time) bool {
	if startTime.IsZero() || endTime.IsZero() {
		return true
	}
	a := now.Sub(startTime)
	b := endTime.Sub(now)

	return a >= 0 && b >= 0
}

func submissionSName(method string) string {
	return fmt.Sprintf("%s.SubmissionService.%s", "xmc.srv.core", method)
}

func sendToDispatcher(req *submission.CreateRequest, id, datasetID uuid.UUID, methodName string) error {
	client := job.NewJobsServiceClient("xmc.srv.dispatcher", client.DefaultClient)
	_, err := client.Create(C(), &job.CreateRequest{
		Priority: 1,
		Job: &job.Job{
			DatasetId:    datasetID.String(),
			Code:         req.Code,
			Language:     req.Language,
			SubmissionId: id.String(),
			TaskId:       req.TaskId,
		},
	})
	if err != nil {
		return errors.InternalServerError(methodName, errors.Parse(err.Error()).Error())
	}

	return nil
}

func (*SubmissionService) Create(ctx context.Context, req *submission.CreateRequest, rsp *submission.CreateResponse) error {
	methodName := submissionSName("Create")
	if !perms.HasScope(ctx, "submission") {
		return errors.Forbidden(methodName, "you are not allowed to create submissions")
	}
	switch {
	case req == nil:
		return errors.BadRequest(methodName, "invalid request")
	case req.Code == nil:
		return errors.BadRequest(methodName, "invalid code")
	case !common.IsValidLanguage(req.Language):
		return errors.BadRequest(methodName, "invalid language")
	}

	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return errors.BadRequest(methodName, "invalid task_id")
	}

	dd := db.DB.BeginGroup()
	task, err := dd.ReadTask(taskID)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.BadRequest(methodName, "task doesn't exist")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	var taskList *tasklist.TaskList
	if task.TaskListID != uuid.Nil {
		taskList, err = dd.ReadTaskList(task.TaskListID)
		if err != nil {
			dd.Rollback()
			if err == db.ErrNotFound {
				return errors.BadRequest(methodName, "task list doesn't exist")
			}
			return errors.InternalServerError(methodName, e(err))
		}
	}

	// supreme submission permissions means that you can submit anytime
	if !perms.HasScope(ctx, "manage/submission") {
		if taskList == nil || (taskList.StartTime != nil &&
			taskList.EndTime != nil && !timeInRange(*taskList.StartTime, time.Now(), *taskList.EndTime)) {
			dd.Rollback()
			return errors.Forbidden(methodName, "task is not currently open for submissions")
		}
	}

	u, err := perms.AccountUUIDFromContext(ctx)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, err.Error())
	}
	sb := &submission.Submission{
		TaskId:   req.TaskId,
		Language: req.Language,
		UserId:   u.String(),
	}
	sb.CreatedAt, _ = ptypes.TimestampProto(time.Now())

	id, err := dd.CreateSubmission(sb)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	attID, err := util.MakeAttachment(dd, &attachment.CreateRequest{
		Attachment: &attachment.Attachment{
			ObjectId: "submission/" + id.String(),
			Filename: "submission." + req.Language,
		},
		Contents: req.Code,
	})
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.SubmissionSetAttachmentID(id, attID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.SubmissionSetDatasetID(id, task.DatasetID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	err = sendToDispatcher(req, id, task.DatasetID, methodName)
	if err != nil {
		return err
	}

	rsp.Id = id.String()
	return nil
}

func (*SubmissionService) Read(ctx context.Context, req *submission.ReadRequest, rsp *submission.ReadResponse) error {
	methodName := submissionSName("Read")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	var sub *submission.Submission
	var res *result.Result
	var ts []*result.TestResult
	s, err := dd.ReadSubmission(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "submission not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	accID, err := perms.AccountUUIDFromContext(ctx)
	if !perms.HasScope(ctx, "manage/submission") && (err != nil || s.UserID != accID) {
		dd.Rollback()
		return errors.Forbidden(methodName, "you are not allowed to read this submission")
	}
	if req.IncludeTestResults {
		trs, err := dd.ReadTestResults(id)
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
		if len(trs) > 0 {
			for _, tr := range trs {
				ts = append(ts, tr.ToProto())
			}
		}
	}
	if req.IncludeResult {
		r, err := dd.ReadSubmissionResult(id)
		if err != nil && err != db.ErrNotFound {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		} else if err == nil {
			res = r.ToProto(ts)
		}
	}
	if res == nil && ts != nil {
		res = &result.Result{
			TestResults: ts,
		}
	}
	sub = s.ToProto(res)

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Submission = sub
	return nil
}

func (*SubmissionService) Update(ctx context.Context, req *submission.UpdateRequest, rsp *submission.UpdateResponse) error {
	methodName := submissionSName("Update")

	if req.Job == nil {
		return errors.BadRequest(methodName, "invalid job")
	}
	if !perms.HasScope(ctx, "manage/submission") {
		return errors.Forbidden(methodName, "you are not allowed to update this submission")
	}

	err := db.DB.UpdateSubmission(req)
	if err != nil {
		if err == db.ErrNotFound {
			log.Warn("Dispatcher tried to update non-existent submission")
		} else {
			return errors.InternalServerError(methodName, e(err))
		}
	}
	return nil
}

func (*SubmissionService) Delete(ctx context.Context, req *submission.DeleteRequest, rsp *submission.DeleteResponse) error {
	methodName := submissionSName("Delete")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}
	if !perms.HasScope(ctx, "manage/submission") {
		return errors.Forbidden(methodName, "you are not allowed to delete this submission")
	}

	dd := db.DB.BeginGroup()
	s, err := dd.ReadSubmission(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, e(err))
		}
		return errors.InternalServerError(methodName, e(err))
	}
	err = dd.DeleteSubmission(id)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	err = util.DeleteAttachment(dd, s.AttachmentID)
	if err != nil {
		dd.Rollback()
		log.Warnf("Error while deleting submission attachment id %v: %v", s.AttachmentID, err)
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (subService *SubmissionService) Search(ctx context.Context, req *submission.SearchRequest, rsp *submission.SearchResponse) error {
	methodName := submissionSName("Search")

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	if len(req.TaskId) > 0 {
		_, err := uuid.Parse(req.TaskId)
		if err != nil {
			return errors.BadRequest(methodName, "invalid task_id")
		}
	}
	if len(req.UserId) > 0 {
		_, err := uuid.Parse(req.UserId)
		if err != nil {
			return errors.BadRequest(methodName, "invalid user_id")
		}
	}

	accID := uuid.Nil
	if !perms.HasScope(ctx, "manage/submission") {
		var err error
		accID, err = perms.AccountUUIDFromContext(ctx)
		if err != nil {
			if err == perms.ErrMissingToken {
				rsp.Submissions = []*submission.Submission{}
				rsp.Meta = &searchmeta.Meta{
					PerPage: req.Limit,
					Count:   0,
					Total:   0,
				}
				return nil
			}
			return errors.InternalServerError(methodName, e(err))
		}
	}

	dd := db.DB.BeginGroup()
	ss, cnt, err := dd.SearchSubmission(req, accID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	subs := []*submission.Submission{}
	for _, s := range ss {
		var r *result.Result
		var ts []*result.TestResult
		if req.IncludeTestResults {
			trs, err := dd.ReadTestResults(s.ID)
			if err != nil {
				dd.Rollback()
				return errors.InternalServerError(methodName, e(err))
			}
			ts = []*result.TestResult{}
			for _, tr := range trs {
				ts = append(ts, tr.ToProto())
			}
		}
		if req.IncludeResult {
			res, err := dd.ReadSubmissionResult(s.ID)
			if err != nil && err != db.ErrNotFound {
				dd.Rollback()
				return errors.InternalServerError(methodName, e(err))
			} else if err == nil {
				r = res.ToProto(ts)
			}
		}
		if req.IncludeTestResults && r == nil {
			r = &result.Result{
				TestResults: ts,
			}
		}
		sub := s.ToProto(r)
		subs = append(subs, sub)
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Submissions = subs
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(subs)),
		Total:   cnt,
	}
	return nil
}
