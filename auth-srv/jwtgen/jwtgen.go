package jwtgen

import (
	"crypto/rsa"

	"github.com/RangelReale/osin"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/auth-srv/account"
	"github.com/xmc-dev/xmc/auth-srv/consts"
	"github.com/xmc-dev/xmc/auth-srv/role"
	"github.com/xmc-dev/xmc/auth-srv/storage"
)

// JWTAccessTokenGen is an implementation of the
// osin.AccessTokenGen interface for JSON Web Tokens
type JWTAccessTokenGen struct {
	PrivateKey *rsa.PrivateKey
}

func (gen *JWTAccessTokenGen) GenerateAccessToken(data *osin.AccessData, generateRefresh bool) (accessToken string, refreshToken string, err error) {
	rawUserData := data.UserData
	claims := jwt.MapClaims{
		"iss":   consts.ServiceName,
		"exp":   data.ExpireAt().Unix(),
		"iat":   data.CreatedAt.Unix(),
		"scope": data.Scope,
		"aud":   data.Client.GetId(),
	}
	if rawUserData != nil {
		claims["sub"] = rawUserData.(*storage.UserData).Subject
		acc, err := account.GetClient(claims["sub"].(string))
		if err != nil {
			return "", "", err
		}
		r, e := role.ReadRole(acc.RoleId)
		if e != nil {
			return "", "", e
		}
		claims["role"] = jwt.MapClaims{
			"id":    r.Id,
			"name":  r.Name,
			"scope": r.Scope,
		}
	} else {
		claims["sub"] = data.Client.GetId()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	accessToken, err = token.SignedString(gen.PrivateKey)
	if err != nil {
		return "", "", err
	}

	if generateRefresh {
		refreshToken = uuid.New().String()
	}

	return
}

// NewJWTAccessTokenGen returns a new instance
func NewJWTAccessTokenGen(privKey *rsa.PrivateKey) *JWTAccessTokenGen {
	return &JWTAccessTokenGen{PrivateKey: privKey}
}
