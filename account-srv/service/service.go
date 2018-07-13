package service

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/micro/cli"
	micro "github.com/micro/go-micro"
	"github.com/xmc-dev/registry"
	"github.com/xmc-dev/xmc/account-srv/consts"
	"github.com/xmc-dev/xmc/common/config"
	"github.com/xmc-dev/xmc/common/perms"
)

// Service represents the micro service along with its config options
type Service struct {
	Micro  micro.Service
	DBType string
	DBURL  string
	Debug  bool

	Consul *consul.Client
}

var MainService *Service

// NewService returns a service definition
func NewService() *Service {
	s := Service{}
	reg := xmcconsul.NewRegistry().(*xmcconsul.XMCConsulRegistry)
	s.Consul = reg.Client
	confReader := config.NewReader(s.Consul, "xmc.srv.account/cfg")
	confReader.MustReadConfig()
	s.Micro = micro.NewService(
		micro.Name(consts.ServiceName),
		micro.WrapHandler(perms.JWTWrapper),
		//micro.WrapHandler(logWrapper),
		micro.Flags(
			cli.StringFlag{
				Name:        "database_url",
				EnvVar:      "CFG_DB_URL",
				Usage:       "The database URL and its parameters e.g. user:password@/dbname?parseTime=True",
				Destination: &s.DBURL,
			},
			cli.StringFlag{
				Name:        "database_type",
				Value:       "postgres",
				Usage:       "The database type e.g. \"mysql\", \"postgres\", \"sqlite3\"",
				Destination: &s.DBType,
			},
			cli.BoolFlag{
				Name:        "debug",
				EnvVar:      "DEBUG",
				Usage:       "Enable debug messages",
				Destination: &s.Debug,
			},
		),
	)

	return &s
}

//func logWrapper(fn server.HandlerFunc) server.HandlerFunc {
//	return func(ctx context.Context, req server.Request, rsp interface{}) error {
//		fmt.Fprintf(os.Stderr, "[%v] server request: %#v\n", time.Now(), req.Request())
//		return fn(ctx, req, rsp)
//	}
//}
