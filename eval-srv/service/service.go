package service

import (
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/xmc-dev/xmc/common/config"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/eval-srv/consts"
	"github.com/xmc-dev/registry"
)

// Service represents the micro service along with its config options
type Service struct {
	Micro  micro.Service
	Consul *consul.Client

	Name        string
	Description string

	OAuth2Token string

	Debug bool
}

// MainService is the service object used throughout the codebase
var MainService *Service

// NewService returns a service definition
func NewService() *Service {
	s := new(Service)
	reg := xmcconsul.NewRegistry().(*xmcconsul.XMCConsulRegistry)
	s.Consul = reg.Client
	confReader := config.NewReader(s.Consul, "xmc.srv.eval/cfg", "xmc/cfg")
	confReader.MustReadConfig()
	registry.DefaultRegistry = reg

	hname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	s.Micro = micro.NewService(
		micro.Name(consts.ServiceName),
		micro.Registry(reg),
		micro.WrapHandler(perms.JWTWrapper),

		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(15*time.Second),

		micro.Flags(
			cli.StringFlag{
				Name:        "name",
				Usage:       "Set the name of the eval node",
				Value:       hname + "-" + strings.Split(uuid.New().String(), "-")[1],
				Destination: &s.Name,
			},
			cli.StringFlag{
				Name:        "description",
				Usage:       "Set the description of the eval node",
				Destination: &s.Description,
			},
			cli.BoolFlag{
				Name:        "debug",
				EnvVar:      "DEBUG",
				Usage:       "Enable debug messages",
				Destination: &s.Debug,
			},
			cli.StringFlag{
				Name:        "oauth2_token",
				Usage:       "API endpoint for the token",
				EnvVar:      "CFG_TOKEN",
				Destination: &s.OAuth2Token,
			},
		),
	)

	return s
}
