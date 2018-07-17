package main

import (
	"crypto/rsa"
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	mlog "github.com/micro/go-log"
	"github.com/sirupsen/logrus"
	micrologrus "github.com/tudurom/micro-logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"github.com/xmc-dev/xmc/api-srv/auth"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/handler/account"
	"github.com/xmc-dev/xmc/api-srv/handler/attachment"
	"github.com/xmc-dev/xmc/api-srv/handler/dataset"
	"github.com/xmc-dev/xmc/api-srv/handler/grader"
	"github.com/xmc-dev/xmc/api-srv/handler/page"
	"github.com/xmc-dev/xmc/api-srv/handler/role"
	"github.com/xmc-dev/xmc/api-srv/handler/submission"
	"github.com/xmc-dev/xmc/api-srv/handler/task"
	"github.com/xmc-dev/xmc/api-srv/handler/tasklist"
	"github.com/xmc-dev/xmc/api-srv/service"
)

var log = logrus.WithField("prefix", "main")

func main() {
	gin.SetMode(gin.ReleaseMode)
	service.MainService = service.NewService()
	srv := service.MainService

	logrus.SetFormatter(new(prefixed.TextFormatter))
	ml := micrologrus.NewMicroLogrus(logrus.StandardLogger().WithField("prefix", "micro"))
	mlog.SetLogger(ml)
	srv.Router.Use(ginrus.Ginrus(logrus.StandardLogger().WithField("prefix", "gin"), time.RFC3339, true))
	srv.Web.Init()
	log.Info("Initialized micro")

	if srv.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	mountMiddleware()
	register("/submissions", &submission.Handler{})
	register("/tasks", &task.Handler{})
	register("/attachments", &attachment.Handler{})
	register("/datasets", &dataset.Handler{})
	register("/graders", &grader.Handler{})
	register("/pages", &page.Handler{})
	register("/pageversions", &page.VersionHandler{})
	register("/pagechildren", &page.ChildrenHandler{})
	register("/accounts", &account.Handler{})
	register("/tasklists", &tasklist.Handler{})
	register("/roles", &role.Handler{})

	if err := srv.Web.Run(); err != nil {
		log.Fatal("Couldn't run service: ", err)
	}
}

func register(path string, h handler.Handler) {
	handler.Register(service.MainService.Router.Group(path), h)
}

func getPubkey() (*rsa.PublicKey, error) {
	c := service.MainService.Consul
	kv, _, err := c.KV().Get("xmc.srv.auth/pubkey", nil)
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, errors.New("No pubkey in consul")
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(kv.Value)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func mountMiddleware() {
	key, err := getPubkey()
	if err != nil {
		log.Fatal("Couldn't get public key: ", err)
	}
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "authorization")
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "HEAD", "PATCH", "DELETE"}
	service.MainService.Router.Use(cors.New(corsConfig), auth.Auth(key))
}
