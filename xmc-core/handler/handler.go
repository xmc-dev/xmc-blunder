package handler

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/common/cred"
	"github.com/xmc-dev/xmc/common/req"
	"github.com/xmc-dev/xmc/xmc-core/service"
)

var log = logrus.WithField("prefix", "handler")

func e(err error) string {
	return fmt.Sprintf("%+v", err)
}

func InitAuth() error {
	return cred.InitAuth("xmc.srv.core", service.MainService.Consul.KV(), service.MainService.OAuth2Token,
		"xmc.dispatcher/create")
}

func C() context.Context {
	ctx, err := req.C(cred.Src)
	if err != nil {
		logrus.WithError(err).Error("Couldn't get token from source")
		return context.Background()
	}

	return ctx
}
