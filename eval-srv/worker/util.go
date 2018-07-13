package worker

import (
	"context"
	"os/exec"
	"strings"

	"github.com/micro/go-micro/metadata"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/common/cred"
	"github.com/xmc-dev/xmc/common/req"
	"github.com/xmc-dev/xmc/eval-srv/service"
)

func cmdString(cmd *exec.Cmd) string {
	s := cmd.Path
	if len(cmd.Args) > 1 {
		s += " " + strings.Join(cmd.Args[1:], " ")
	}

	return s
}

func InitAuth() error {
	return cred.InitAuth("xmc.srv.eval", service.MainService.Consul.KV(), service.MainService.OAuth2Token,
		"xmc.core/manage/attachment xmc.dispatcher/finish")
}

func CWithName(evalName string) context.Context {
	ctx, err := req.C(cred.Src, metadata.Metadata{"X-Eval-Name": evalName})
	if err != nil {
		logrus.WithError(err).Error("Couldn't get token from source")
		return context.Background()
	}

	return ctx
}

func C() context.Context {
	ctx, err := req.C(cred.Src)
	if err != nil {
		logrus.WithError(err).Error("Couldn't get token from source")
		return context.Background()
	}

	return ctx
}
