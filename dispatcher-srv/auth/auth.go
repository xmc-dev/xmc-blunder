package auth

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/common/cred"
	"github.com/xmc-dev/xmc/common/req"
	"github.com/xmc-dev/xmc/dispatcher-srv/service"
)

func InitAuth() error {
	return cred.InitAuth("xmc.srv.dispatcher", service.MainService.Consul.KV(), service.MainService.OAuth2Token,
		"xmc.core/manage/submission xmc.eval/assign")
}

func C() context.Context {
	ctx, err := req.C(cred.Src)
	if err != nil {
		logrus.WithError(err).Error("Couldn't get token from source")
		return context.Background()
	}

	return ctx
}
