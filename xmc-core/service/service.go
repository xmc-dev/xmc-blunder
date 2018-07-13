package service

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/micro/cli"
	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/xmc-dev/xmc/common/config"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/registry"
)

// Service represents the micro service along with its config options
type Service struct {
	Micro  micro.Service
	Consul *consul.Client

	DBType string
	DBURL  string

	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3BucketName      string
	S3BucketLocation  string
	S3UseSSL          bool

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
	confReader := config.NewReader(s.Consul, "xmc.srv.core/cfg")
	confReader.MustReadConfig()
	registry.DefaultRegistry = reg
	s.Micro = micro.NewService(
		micro.Name("xmc.srv.core"),
		micro.Registry(reg),

		micro.WrapHandler(perms.JWTWrapper),

		micro.Flags(
			cli.BoolFlag{
				Name:        "debug",
				EnvVar:      "DEBUG",
				Usage:       "Enable debug messages",
				Destination: &s.Debug,
			},
			cli.StringFlag{
				Name:        "database_url",
				EnvVar:      "CFG_DB_URL",
				Usage:       "The database URL and its parameters e.g. user:password@/dbname?parseTime=True",
				Destination: &s.DBURL,
			},
			cli.StringFlag{
				Name:        "database_type",
				Value:       "postgres",
				Usage:       "The database type (currently only postgres)",
				Destination: &s.DBType,
			},
			cli.StringFlag{
				Name:        "s3_endpoint",
				Value:       "s3.amazonaws.com",
				Usage:       "The endpoint of the S3 server",
				EnvVar:      "CFG_S3_ENDPOINT",
				Destination: &s.S3Endpoint,
			},
			cli.StringFlag{
				Name:        "s3_access_key_id",
				EnvVar:      "CFG_S3_ACCESS_KEY_ID",
				Usage:       "The S3 access key id",
				Destination: &s.S3AccessKeyID,
			},
			cli.StringFlag{
				Name:        "s3_secret_access_key",
				EnvVar:      "CFG_S3_SECRET_ACCESS_KEY",
				Usage:       "The S3 secret access key",
				Destination: &s.S3SecretAccessKey,
			},
			cli.BoolTFlag{
				Name:        "s3_use_ssl",
				Usage:       "Whether to to use SSL when connecting to S3",
				EnvVar:      "CFG_S3_USE_SSL",
				Destination: &s.S3UseSSL,
			},
			cli.StringFlag{
				Name:        "s3_bucket_name",
				Usage:       "The name of the S3 bucket",
				EnvVar:      "CFG_S3_BUCKET_NAME",
				Destination: &s.S3BucketName,
			},
			cli.StringFlag{
				Name:        "s3_bucket_location",
				Usage:       "The location of the S3 bucket",
				Value:       "eu-central-1",
				EnvVar:      "CFG_S3_BUCKET_LOCATION",
				Destination: &s.S3BucketLocation,
			},
			cli.StringFlag{
				Name:        "oauth2_token",
				Usage:       "The API endpoint for tokens",
				EnvVar:      "CFG_TOKEN",
				Destination: &s.OAuth2Token,
			},
		),
	)

	return s
}
