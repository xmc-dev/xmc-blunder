package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/dispatcher-srv/auth"
	"github.com/xmc-dev/xmc/dispatcher-srv/consts"
	"github.com/xmc-dev/xmc/dispatcher-srv/db"
	dbjob "github.com/xmc-dev/xmc/dispatcher-srv/db/models/job"
	"github.com/xmc-dev/xmc/dispatcher-srv/dispatch"
	"github.com/xmc-dev/xmc/dispatcher-srv/proto/job"
	"github.com/xmc-dev/xmc/xmc-core/proto/submission"
)

var log = logrus.WithField("prefix", "job_handler")
var submissionClient = submission.NewSubmissionServiceClient("xmc.srv.core", client.DefaultClient)

// JobsService manages job operations
type JobsService struct{}

func jobSName(method string) string {
	return fmt.Sprintf("%s.JobsService.%s", consts.ServiceName, method)
}

func (*JobsService) Create(ctx context.Context, req *job.CreateRequest, rsp *job.CreateResponse) error {
	methodName := jobSName("Create")
	j := req.Job

	switch {
	case req.Job == nil:
		return errors.BadRequest(methodName, "invalid job")
	case req.Priority <= 0:
		return errors.BadRequest(methodName, "invalid priority")
	case len(j.DatasetId) == 0:
		return errors.BadRequest(methodName, "invalid dataset_id")
	case len(j.Code) == 0:
		return errors.BadRequest(methodName, "invalid code")
	case len(j.Language) == 0:
		return errors.BadRequest(methodName, "invalid language")
	case len(j.SubmissionId) == 0:
		return errors.BadRequest(methodName, "invalid submission_id")
	case len(j.TaskId) == 0:
		return errors.BadRequest(methodName, "invalid task_id")
	}

	if !perms.HasScope(ctx, "create") {
		return errors.Forbidden(methodName, "you are not allowed to create jobs")
	}

	j.Result = nil
	j.EvalId = ""
	j.State = job.State_WAITING
	j.CreatedAt, _ = ptypes.TimestampProto(time.Time{})
	j.FinishedAt, _ = ptypes.TimestampProto(time.Time{})
	u, err := db.CreateJob(j)
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}
	j.Uuid = u.String()

	jj, err := db.ReadJob(u.String())
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}
	_, err = submissionClient.Update(auth.C(), &submission.UpdateRequest{
		Job: jj.ToProto(),
	})
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}

	err = db.EnqueueJob(int(req.Priority), u)
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}
	go func() {
		qi, err := db.GetFirstJobInQueue()
		if err != nil {
			log.WithFields(logrus.Fields{
				"req": req,
				"qi":  qi,
				"err": err,
			}).Error("couldn't get first job in queue in Create")
		}
		log.WithFields(logrus.Fields{
			"first_in_queue": qi.JobUUID,
			"current":        j.Uuid,
		}).Debug("First job in queue vs current job")
		if qi.JobUUID.String() == j.Uuid {
			dispatch.Next()
		}
	}()
	rsp.Uuid = u.String()

	return nil
}

func (*JobsService) Read(ctx context.Context, req *job.ReadRequest, rsp *job.ReadResponse) error {
	methodName := jobSName("Read")
	if len(req.Uuid) == 0 {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	j, err := db.ReadJob(req.Uuid)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "job not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	rsp.Job = j.ToProto()
	return nil
}

func (*JobsService) Search(ctx context.Context, req *job.SearchRequest, rsp *job.SearchResponse) error {
	methodName := jobSName("Search")
	if req.Limit == 0 {
		req.Limit = 10
	}

	jobs, err := db.SearchJob(req)
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}

	for _, j := range jobs {
		rsp.Jobs = append(rsp.Jobs, j.ToProto())
	}

	return nil
}

func (*JobsService) Finish(ctx context.Context, req *job.FinishRequest, rsp *job.FinishResponse) error {
	methodName := jobSName("Finish")
	if _, err := uuid.Parse(req.JobUuid); err != nil {
		return errors.BadRequest(methodName, "invalid job_uuid")
	}
	if req.Result == nil {
		return errors.BadRequest(methodName, "invalid result")
	}
	if db.IsFinished(req.JobUuid) {
		return errors.BadRequest(methodName, "job is finished")
	}

	if !perms.HasScope(ctx, "finish") {
		return errors.Forbidden(methodName, "you are not allowed to finish jobs")
	}

	err := db.FinishJob(req)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "job not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	j, err := db.ReadJob(req.JobUuid)
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}
	pj := j.ToProto()
	pj.Result = req.Result
	_, err = submissionClient.Update(auth.C(), &submission.UpdateRequest{Job: pj})
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}
	qi, err := db.DequeueJob()
	if err == db.ErrNotFound {
		rsp.NextJob = nil
	} else if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	} else {
		j, err := db.ReadJob(qi.JobUUID.String())
		if err != nil {
			return errors.InternalServerError(methodName, err.Error())
		}
		// if the context doesn't have metadata then it's a problem with micro
		meta, _ := metadata.FromContext(ctx)
		// and if there's no x-eval-name, the server will panic.
		// i will read the stack trace and punish myself somehow for letting that happen
		evalName := meta["X-Eval-Name"]
		err = db.SetJobStateAndEvalID(j.UUID, dbjob.PROCESSING, evalName)
		if err != nil {
			return errors.InternalServerError(methodName, err.Error())
		}
		rsp.NextJob = j.ToProto()
	}
	return nil
}
