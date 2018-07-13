package main

import (
	"fmt"
	"time"

	mlog "github.com/micro/go-log"
	"github.com/sirupsen/logrus"
	micrologrus "github.com/tudurom/micro-logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/common/wait"
	"github.com/xmc-dev/xmc/dispatcher-srv/auth"
	"github.com/xmc-dev/xmc/dispatcher-srv/db"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/sql"
	"github.com/xmc-dev/xmc/dispatcher-srv/dispatch"
	"github.com/xmc-dev/xmc/dispatcher-srv/handler"
	"github.com/xmc-dev/xmc/dispatcher-srv/proto/job"
	"github.com/xmc-dev/xmc/dispatcher-srv/proto/meta"
	"github.com/xmc-dev/xmc/dispatcher-srv/service"
	"github.com/xmc-dev/xmc/dispatcher-srv/status"
)

var srv *service.Service
var log = logrus.WithField("prefix", "main")
var nis []*status.NodeInfo

var firstCheckDone = make(chan bool)

func sendEv() {
	t := time.NewTicker(time.Duration(srv.HealthCheckInterval) * time.Second)

	for ; true; <-t.C {
		nis = status.HealthCheck()
		if len(nis) == 0 {
			log.Warn("No node is alive")
		} else {
			firstCheckDone <- true
		}
	}
}

func kickstartDispatch() {
	fmt.Println("!!waiting for first check")
	<-firstCheckDone
	fmt.Println("!!first check done")
	dispatch.Next()
}

func main() {
	service.MainService = service.NewService()
	srv = service.MainService

	perms.Register(rawPerms, treeRoot)
	logrus.SetFormatter(new(prefixed.TextFormatter))
	ml := micrologrus.NewMicroLogrus(logrus.StandardLogger().WithField("prefix", "micro"))
	mlog.SetLogger(ml)
	srv.Micro.Init()

	if err := wait.For("xmc.srv.auth", "xmc.srv.eval", "xmc.srv.core"); err != nil {
		log.WithError(err).Fatal("Couldn't wait")
	}
	if err := perms.GetPubkey(service.MainService.Consul.KV()); err != nil {
		log.WithError(err).Fatal("Couldn't get pubkey")
	}
	if err := auth.InitAuth(); err != nil {
		log.WithError(err).Fatal("Couldn't init auth")
	}

	if srv.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	job.RegisterJobsServiceHandler(srv.Micro.Server(), &handler.JobsService{})
	meta.RegisterMetaServiceHandler(srv.Micro.Server(), &handler.MetaService{&nis})

	dbLog := log.WithFields(logrus.Fields{
		"database_url":  srv.DBURL,
		"database_type": srv.DBType,
	})
	if srv.DBURL == "" {
		dbLog.Fatal("Invalid DB URL")
	}
	db.Register(&sql.SQL{})
	if err := db.Init(srv); err != nil {
		dbLog.Fatal("Couldn't initialize database: ", err)
	}
	defer db.Deinit()

	status.InitHealthCheck()
	// Try to dispatch something to get the flow started
	go sendEv()
	go kickstartDispatch()

	if err := srv.Micro.Run(); err != nil {
		logrus.Fatal("Couldn't run service: ", err)
	}
}
