package main

import (
	mlog "github.com/micro/go-log"
	"github.com/sirupsen/logrus"
	micrologrus "github.com/tudurom/micro-logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/common/wait"
	"github.com/xmc-dev/xmc/eval-srv/handler"
	"github.com/xmc-dev/xmc/eval-srv/proto/eval"
	"github.com/xmc-dev/xmc/eval-srv/service"
	"github.com/xmc-dev/xmc/eval-srv/worker"
)

var srv *service.Service

var log = logrus.WithField("prefix", "main")

func main() {
	service.MainService = service.NewService()
	srv = service.MainService
	handler.InitHandler()

	perms.Register(rawPerms, treeRoot)
	logrus.SetFormatter(new(prefixed.TextFormatter))
	ml := micrologrus.NewMicroLogrus(logrus.StandardLogger().WithField("prefix", "micro"))
	mlog.SetLogger(ml)
	srv.Micro.Init()
	log.WithFields(logrus.Fields{
		"name": srv.Name,
		"desc": srv.Description,
	}).Info("Initialized micro")

	if err := wait.For("xmc.srv.auth"); err != nil {
		log.WithError(err).Fatal("Couldn't wait")
	}
	if err := perms.GetPubkey(service.MainService.Consul.KV()); err != nil {
		log.WithError(err).Fatal("Couldn't get pubkey")
	}

	if err := worker.InitAuth(); err != nil {
		log.WithError(err).Fatal("Couldn't init auth")
	}

	if srv.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	eval.RegisterEvalServiceHandler(srv.Micro.Server(), &handler.EvalService{Worker: worker.NewWorker(service.MainService)})

	if err := srv.Micro.Run(); err != nil {
		log.Fatal("Couldn't run service: ", err)
	}
}
