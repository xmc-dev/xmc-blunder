package handler

import (
	"context"
	"fmt"

	"github.com/micro/go-micro/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/job"
	"github.com/xmc-dev/xmc/eval-srv/consts"
	"github.com/xmc-dev/xmc/eval-srv/proto/eval"
	"github.com/xmc-dev/xmc/eval-srv/worker"
)

type EvalService struct {
	Worker *worker.Worker

	disabled bool
}

func evalSName(method string) string {
	return fmt.Sprintf("%s.EvalService.%s", consts.ServiceName, method)
}

func (es *EvalService) Assign(ctx context.Context, req *eval.AssignRequest, rsp *eval.AssignResponse) error {
	methodName := evalSName("Assign")

	switch {
	case es.disabled:
		return errors.BadRequest(methodName, "node is disabled")
	case req.Job == nil:
		return errors.BadRequest(methodName, "invalid job")
	}

	if !perms.HasScope(ctx, "assign") {
		return errors.Forbidden(methodName, "you are not allowed to assign submissions")
	}

	j := req.Job
	jb := job.FromProto(j)
	err := es.Worker.Work(jb)
	if err != nil {
		return errors.BadRequest(methodName, err.Error())
	}

	return nil
}

func (es *EvalService) GetStatus(ctx context.Context, req *eval.GetStatusRequest, rsp *eval.GetStatusResponse) error {
	rsp.Info = &eval.NodeInfo{
		Id:          srv.Micro.Server().Options().Id,
		Name:        srv.Name,
		Description: srv.Description,
		Idle:        es.Worker.IsIdle(),
		Disabled:    es.disabled,
	}
	return nil
}

func (es *EvalService) SetDisabled(ctx context.Context, req *eval.SetDisabledRequest, rsp *eval.SetDisabledResponse) error {
	es.disabled = req.Disabled
	if es.disabled {
		logrus.Warn("Node has been disabled!")
	} else {
		logrus.Warn("Node has been enabled!")
	}

	return nil
}
