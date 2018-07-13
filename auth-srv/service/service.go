package service

import (
	"errors"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	consul "github.com/hashicorp/consul/api"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/micro/cli"
	web "github.com/micro/go-web"
	"github.com/urfave/negroni"
	"github.com/xmc-dev/registry"
	"github.com/xmc-dev/xmc/auth-srv/consts"
	"github.com/xmc-dev/xmc/auth-srv/globals"
	"github.com/xmc-dev/xmc/auth-srv/handlers/authorize"
	"github.com/xmc-dev/xmc/auth-srv/handlers/certs"
	"github.com/xmc-dev/xmc/auth-srv/handlers/login"
	"github.com/xmc-dev/xmc/auth-srv/handlers/logout"
	"github.com/xmc-dev/xmc/auth-srv/handlers/token"
	"github.com/xmc-dev/xmc/common/config"
)

type Service struct {
	Micro         web.Service
	Consul        *consul.Client
	DBUrl         string
	Debug         bool
	StoragePrefix string
	ConsentPrefix string

	router *mux.Router
}

var MainService *Service

func getRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	// /
	r.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprintln(w, "hello from xmc.srv.auth")
	})

	// /login
	r.StrictSlash(false).HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.RawQuery
		redir := "/login/"
		if len(q) > 0 {
			redir += "?" + q
		}
		http.Redirect(w, r, redir, http.StatusMovedPermanently)
	}).Methods("GET")
	r.StrictSlash(false).Path("/login/").Handler(
		login.CheckLoginHandler(
			http.HandlerFunc(
				login.GETHandler))).Methods("GET")
	r.Path("/login").Handler(
		http.HandlerFunc(
			login.POSTHandler)).Methods("POST")
	r.PathPrefix("/login/").Handler(
		http.StripPrefix("/login/",
			http.FileServer(http.Dir("./static/login")))).Methods("GET")

	// /logout
	r.HandleFunc("/logout", logout.Handler)

	// /authorize
	r.HandleFunc("/authorize", authorize.Handler)

	// /token
	r.HandleFunc("/token", token.Handler)

	// /certs
	r.HandleFunc("/certs", certs.Handler)
	return r
}

func NewService() *Service {
	s := Service{}
	s.router = getRouter()
	n := negroni.New()
	n.Use(negronilogrus.NewMiddleware())
	n.Use(NoCache{})
	n.UseHandler(handlers.CORS()(s.router))
	registry := xmcconsul.NewRegistry().(*xmcconsul.XMCConsulRegistry)
	s.Consul = registry.Client
	cfgReader := config.NewReader(s.Consul, "xmc.srv.auth/cfg")
	cfgReader.MustReadConfig()
	s.Micro = web.NewService(
		web.Name(consts.ServiceName),
		web.Version("latest"),
		web.Handler(n),
		web.Registry(registry),
		web.Metadata(map[string]string{
			"RAW-traefik.enable":               "true",
			"RAW-traefik.frontend.entryPoints": "http",
			"RAW-traefik.frontend.rule":        "PathPrefixStrip:/oauth2/",
		}),
		web.Flags(
			cli.StringFlag{
				Name:        "database_url",
				EnvVar:      "CFG_DB_URL",
				Usage:       "The redis database URL",
				Value:       ":6379",
				Destination: &s.DBUrl,
			},
			cli.BoolFlag{
				Name:        "debug",
				EnvVar:      "DEBUG",
				Usage:       "Enable debug messages",
				Destination: &s.Debug,
			},
			cli.StringFlag{
				Name:        "storage_prefix",
				EnvVar:      "CFG_STORAGE_PREFIX",
				Usage:       "Prefix for Oauth2 token storage in the redis db",
				Value:       "auth-srv-oauth2",
				Destination: &s.StoragePrefix,
			},
			cli.StringFlag{
				Name:        "consent_prefix",
				EnvVar:      "CFG_CONSENT_PREFIX",
				Usage:       "Prefix for user consent request storage in the redis db",
				Value:       "auth-srv-consent",
				Destination: &s.ConsentPrefix,
			},
		),
	)

	return &s
}

func getKV(c *consul.Client, path string) (*consul.KVPair, error) {
	kv, _, err := c.KV().Get(path, nil)
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, errors.New("No " + path + " in consul")
	}

	return kv, nil
}

func (s *Service) GetKeys() error {
	c := s.Consul
	kv, err := getKV(c, "xmc.srv.auth/privkey")
	if err != nil {
		return err
	}
	globals.PrivKey, err = jwt.ParseRSAPrivateKeyFromPEM(kv.Value)
	if err != nil {
		return err
	}

	kv, err = getKV(c, "xmc.srv.auth/pubkey")
	if err != nil {
		return err
	}
	globals.PubKey, err = jwt.ParseRSAPublicKeyFromPEM(kv.Value)
	if err != nil {
		return err
	}

	return nil
}
