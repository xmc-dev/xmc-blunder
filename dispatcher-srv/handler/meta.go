package handler

import (
	"context"

	"github.com/xmc-dev/xmc/dispatcher-srv/dispatch"
	"github.com/xmc-dev/xmc/dispatcher-srv/proto/meta"
	"github.com/xmc-dev/xmc/dispatcher-srv/status"
	"github.com/xmc-dev/xmc/eval-srv/proto/eval"
)

type MetaService struct {
	NodeInfos *[]*status.NodeInfo
}

func (ms *MetaService) GetEvals(ctx context.Context, req *meta.GetEvalsRequest, rsp *meta.GetEvalsResponse) error {
	if req.Refresh {
		*ms.NodeInfos = status.HealthCheck()
	}
	pnis := []*eval.NodeInfo{}
	for _, ni := range *ms.NodeInfos {
		pnis = append(pnis, ni.ToProto())
	}

	rsp.Evals = pnis
	return nil
}

func (*MetaService) DispatchNext(ctx context.Context, req *meta.DispatchNextRequest, rsp *meta.DispatchNextResponse) error {
	dispatch.Next()
	return nil
}
