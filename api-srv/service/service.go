package service

import (
	"time"

	"github.com/xmc-dev/registry"

	"github.com/gin-gonic/gin"
	consul "github.com/hashicorp/consul/api"
	"github.com/micro/cli"
	web "github.com/micro/go-web"
)

// Service represents the micro go-web service along with its config options
type Service struct {
	Web    web.Service
	Router *gin.Engine
	Consul *consul.Client

	Debug bool
}

var MainService *Service

func NewService() *Service {
	s := new(Service)
	s.Router = gin.New()
	s.Router.Use(gin.Recovery())
	registry := xmcconsul.NewRegistry().(*xmcconsul.XMCConsulRegistry)
	s.Consul = registry.Client
	s.Web = web.NewService(
		web.Name("xmc.srv.api"),
		web.Handler(s.Router),
		web.Registry(registry),
		web.Metadata(map[string]string{
			"RAW-traefik.enable":               "true",
			"RAW-traefik.frontend.entryPoints": "http",
			"RAW-traefik.frontend.rule":        "PathPrefixStrip:/api/",
		}),

		web.RegisterTTL(20*time.Second),
		web.RegisterInterval(10*time.Second),

		web.Flags(
			cli.BoolFlag{
				Name:        "debug",
				EnvVar:      "DEBUG",
				Usage:       "Enable debug messages",
				Destination: &s.Debug,
			},
		),
	)

	return s
}
