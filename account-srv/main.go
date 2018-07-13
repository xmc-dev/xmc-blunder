package main

import (
	mlog "github.com/micro/go-log"
	"github.com/sirupsen/logrus"
	"github.com/tudurom/micro-logrus"
	"github.com/xmc-dev/xmc/account-srv/db"
	"github.com/xmc-dev/xmc/account-srv/db/sql"
	"github.com/xmc-dev/xmc/account-srv/handler"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/account-srv/proto/role"
	"github.com/xmc-dev/xmc/account-srv/proto/session"
	"github.com/xmc-dev/xmc/account-srv/service"
	"github.com/xmc-dev/xmc/common/perms"
)

var log = logrus.WithField("prefix", "main")

func main() {
	srv := service.NewService()
	service.MainService = srv

	perms.Register(rawPerms, treeRoot)
	ml := micrologrus.NewMicroLogrus(logrus.StandardLogger())
	mlog.SetLogger(ml)
	srv.Micro.Init()

	if err := perms.GetPubkey(service.MainService.Consul.KV()); err != nil {
		log.WithError(err).Fatal("Couldn't get pubkey")
	}

	if srv.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	account.RegisterAccountsServiceHandler(srv.Micro.Server(), new(handler.AccountsService))
	session.RegisterSessionsServiceHandler(srv.Micro.Server(), new(handler.SessionsService))
	role.RegisterRoleServiceHandler(srv.Micro.Server(), new(handler.RoleService))

	dbLog := logrus.WithFields(logrus.Fields{
		"database_url":  srv.DBURL,
		"database_type": srv.DBType,
	})
	if srv.DBURL == "" {
		dbLog.Fatal("Invalid DB URL")
	}
	db.Register(&sql.SQL{})
	if err := db.Init(srv.DBType, srv.DBURL); err != nil {
		dbLog.Fatal("Couldn't initialize database: ", err)
	}
	defer db.Deinit()
	rootID, err := db.CreateRootAccount()
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't create root account")
	}
	logrus.WithField("rootID", rootID).Info("We have a root")

	if err := srv.Micro.Run(); err != nil {
		logrus.Fatal("Couldn't run service: ", err)
	}
}
