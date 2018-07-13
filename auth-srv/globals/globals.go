package globals

import (
	"crypto/rsa"

	"github.com/RangelReale/osin"
	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/account-srv/proto/session"
	"github.com/xmc-dev/xmc/auth-srv/jwtgen"
	"github.com/xmc-dev/xmc/auth-srv/storage"
)

// OsinServer is a magic thing that handles oauth2 stuff
var OsinServer *osin.Server

// RedisConn is the connection to the redis database
var RedisConn redis.Conn

var ConsentStorage *storage.ConsentStorage

var AccountClient account.AccountsServiceClient
var SessionClient session.SessionsServiceClient

var PrivKey *rsa.PrivateKey
var PubKey *rsa.PublicKey

// InitRedisConnection initializes the connection to the redis database
func InitRedisConnection(addr string) {
	var err error
	RedisConn, err = redis.Dial("tcp", addr)
	redisLog := logrus.WithField("address", addr)
	if err != nil {
		redisLog.Fatal(err)
	} else {
		redisLog.Info("Connected to redis")
	}
}

// InitConsentStorage initializes the user consent request storage.
func InitConsentStorage(consentPrefix string) {
	ConsentStorage = storage.NewConsentStorage(consentPrefix, RedisConn)
}

// InitOsinServer intializes oauth2 functionality
func InitOsinServer(storagePrefix string) error {
	config := osin.NewServerConfig()
	config.AccessExpiration = 15000
	config.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	config.AllowedAccessTypes = osin.AllowedAccessType{
		osin.AUTHORIZATION_CODE,
		osin.REFRESH_TOKEN,
		osin.CLIENT_CREDENTIALS,
	}
	config.AllowGetAccessRequest = true
	config.RequirePKCEForPublicClients = true
	config.AllowClientSecretInParams = true

	s := storage.NewXMCStorage(storagePrefix, RedisConn)

	OsinServer = osin.NewServer(config, s)
	logrus.Info("Initialized osin server")

	OsinServer.AccessTokenGen = jwtgen.NewJWTAccessTokenGen(PrivKey)
	return nil
}
