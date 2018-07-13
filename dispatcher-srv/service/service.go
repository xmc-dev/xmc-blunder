package service

import (
	"os"

	consul "github.com/hashicorp/consul/api"
	"github.com/micro/cli"
	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	xmcconsul "github.com/xmc-dev/registry"
	"github.com/xmc-dev/xmc/common/config"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/dispatcher-srv/consts"
)

type Service struct {
	Micro               micro.Service
	Debug               bool
	HealthCheckInterval int

	DBType string
	DBURL  string
	DBLog  bool

	OAuth2Token string

	Consul *consul.Client
}

var MainService *Service

func NewService() *Service {
	s := new(Service)
	reg := xmcconsul.NewRegistry().(*xmcconsul.XMCConsulRegistry)
	s.Consul = reg.Client
	confReader := config.NewReader(s.Consul, "xmc.srv.dispatcher/cfg", "xmc/cfg")
	confReader.MustReadConfig()
	registry.DefaultRegistry = reg
	s.Micro = micro.NewService(
		micro.Name(consts.ServiceName),
		micro.WrapHandler(perms.JWTWrapper),
		micro.Registry(reg),

		micro.Flags(
			cli.BoolFlag{
				Name:        "debug",
				EnvVar:      "DEBUG",
				Usage:       "Enable debug messages",
				Destination: &s.Debug,
			},
			cli.IntFlag{
				Name:        "health_check_interval",
				Usage:       "Interval in seconds to execute health checks on dispatchers. Defaults to 10 second",
				Value:       10,
				Destination: &s.HealthCheckInterval,
			},
			cli.StringFlag{
				Name:        "database_url",
				EnvVar:      "CFG_DB_URL",
				Usage:       "The database URL and its parameters e.g. user:password@/dbname?parseTime=True",
				Destination: &s.DBURL}, cli.StringFlag{
				Name:        "database_type",
				Value:       "postgres",
				Usage:       "The database type. Only postgres",
				Destination: &s.DBType,
			},
			cli.BoolFlag{
				Name:        "database_log",
				Usage:       "Enable database logs.",
				Destination: &s.DBLog,
			},
			cli.StringFlag{
				Name:        "oauth2_token",
				EnvVar:      "CFG_TOKEN",
				Usage:       "The API endpoint to get the token",
				Destination: &s.OAuth2Token,
			},
		),
	)

	return s
}

func (s *Service) Stop(exit int) {
	s.Micro.Server().Start()
	os.Exit(exit)
}
