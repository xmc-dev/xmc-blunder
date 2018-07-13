package main

import (
	mlog "github.com/micro/go-log"
	"github.com/sirupsen/logrus"
	"github.com/tudurom/micro-logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/common/wait"
	"github.com/xmc-dev/xmc/xmc-core/db"
	"github.com/xmc-dev/xmc/xmc-core/handler"
	"github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	"github.com/xmc-dev/xmc/xmc-core/proto/dataset"
	"github.com/xmc-dev/xmc/xmc-core/proto/grader"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
	"github.com/xmc-dev/xmc/xmc-core/proto/submission"
	"github.com/xmc-dev/xmc/xmc-core/proto/task"
	"github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
	"github.com/xmc-dev/xmc/xmc-core/s3"
	"github.com/xmc-dev/xmc/xmc-core/service"
)

var srv *service.Service

var log = logrus.WithField("prefix", "main")

func main() {
	service.MainService = service.NewService()
	srv = service.MainService

	perms.Register(rawPerms, treeRoot)
	logrus.SetFormatter(new(prefixed.TextFormatter))
	ml := micrologrus.NewMicroLogrus(logrus.StandardLogger().WithField("prefix", "micro"))
	mlog.SetLogger(ml)
	srv.Micro.Init()
	log.Info("Initialized micro")

	if err := wait.For("xmc.srv.auth"); err != nil {
		log.WithError(err).Fatal("Couldn't wait")
	}
	if err := perms.GetPubkey(service.MainService.Consul.KV()); err != nil {
		log.WithError(err).Fatal("Couldn't get pubkey")
	}
	if err := handler.InitAuth(); err != nil {
		log.WithError(err).Fatal("Couldn't init auth")
	}

	if srv.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	dbLog := log.WithFields(logrus.Fields{
		"database_url": srv.DBURL,
	})
	if srv.DBURL == "" {
		dbLog.Fatal("Invalid DB URL")
	}
	if err := db.Init(srv); err != nil {
		dbLog.Fatal("Couldn't initialize database: ", err)
	}
	defer func() {
		if err := db.Deinit(); err != nil {
			log.Fatal("Couldn't deinit database: ", err)
		}
	}()

	if err := s3.Init(); err != nil {
		log.Fatal("Couldn't connect to S3: ", err)
	}
	defer s3.Deinit()

	attachment.RegisterAttachmentServiceHandler(srv.Micro.Server(), &handler.AttachmentService{})
	dataset.RegisterDatasetServiceHandler(srv.Micro.Server(), &handler.DatasetService{})
	task.RegisterTaskServiceHandler(srv.Micro.Server(), &handler.TaskService{})
	grader.RegisterGraderServiceHandler(srv.Micro.Server(), &handler.GraderService{})
	submission.RegisterSubmissionServiceHandler(srv.Micro.Server(), &handler.SubmissionService{})
	page.RegisterPageServiceHandler(srv.Micro.Server(), &handler.PageService{})
	tasklist.RegisterTaskListServiceHandler(srv.Micro.Server(), &handler.TaskListService{})

	if err := srv.Micro.Run(); err != nil {
		log.Fatal("Couldn't run service: ", err)
	}
}
