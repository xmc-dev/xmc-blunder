package main

import (
	logrus_stack "github.com/Gurpartap/logrus-stack"
	log "github.com/micro/go-log"
	"github.com/micro/go-micro/client"
	"github.com/sirupsen/logrus"
	micrologrus "github.com/tudurom/micro-logrus"
	paccount "github.com/xmc-dev/xmc/account-srv/proto/account"
	prole "github.com/xmc-dev/xmc/account-srv/proto/role"
	psession "github.com/xmc-dev/xmc/account-srv/proto/session"
	"github.com/xmc-dev/xmc/auth-srv/account"
	"github.com/xmc-dev/xmc/auth-srv/globals"
	"github.com/xmc-dev/xmc/auth-srv/role"
	"github.com/xmc-dev/xmc/auth-srv/service"
	"github.com/xmc-dev/xmc/auth-srv/session"
	"github.com/xmc-dev/xmc/auth-srv/storage"
)

func main() {
	srv := service.NewService()
	service.MainService = srv

	ml := micrologrus.NewMicroLogrus(logrus.StandardLogger())
	log.SetLogger(ml)
	if err := srv.Micro.Init(); err != nil {
		logrus.Fatal(err)
	}
	if err := srv.GetKeys(); err != nil {
		logrus.Fatal(err)
	}

	if srv.Debug {
		logrus.AddHook(logrus_stack.StandardHook())
		logrus.SetLevel(logrus.DebugLevel)
	}

	account.Client = paccount.NewAccountsServiceClient("xmc.srv.account", client.DefaultClient)
	session.Client = psession.NewSessionsServiceClient("xmc.srv.account", client.DefaultClient)
	role.Client = prole.NewRoleServiceClient("xmc.srv.account", client.DefaultClient)

	globals.InitRedisConnection(srv.DBUrl)
	globals.InitConsentStorage(srv.ConsentPrefix)
	if err := globals.InitOsinServer(srv.StoragePrefix); err != nil {
		log.Fatal(err)
	}

	if err := srv.Micro.Run(); err != nil {
		log.Fatal(err)
	}

	globals.OsinServer.Storage.(*storage.XMCStorage).CloseDB()
}
