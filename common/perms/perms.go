package perms

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	"github.com/xmc-dev/gandalf"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/common/kv"
)

type ctxKey int

const (
	keyJWT ctxKey = iota
)

var tree *gandalf.ScopeTree
var Pubkey *rsa.PublicKey

var ErrMissingToken = errors.New("missing token")

func Register(yaml, root string) {
	var err error
	tree, err = gandalf.MakeTree([]byte(yaml), root)
	if err != nil {
		panic(err)
	}
}

// ParseToken parses a JWT
func ParseToken(rawToken string) (*jwt.Token, error) {
	return jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}

		return Pubkey, nil
	})
}

// ContextWithToken wraps a JWT in a context
func ContextWithToken(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, keyJWT, token)
}

func TokenFromContext(ctx context.Context) (*jwt.Token, jwt.MapClaims, bool) {
	t, ok := ctx.Value(keyJWT).(*jwt.Token)
	if !ok {
		return nil, nil, false
	}

	return t, t.Claims.(jwt.MapClaims), true
}

func HasScope(ctx context.Context, scope string) bool {
	_, c, ok := TokenFromContext(ctx)
	if !ok {
		return false
	}

	botScope := c["scope"].(string)
	role, hasRole := c["role"]
	if hasRole {
		roleScope := role.(map[string]interface{})["scope"].(string)
		return tree.Need(scope, botScope) && tree.Need(scope, roleScope)
	}
	return tree.Need(scope, botScope)
}

func AccountUUIDFromContext(ctx context.Context) (uuid.UUID, error) {
	_, claims, ok := TokenFromContext(ctx)
	if !ok {
		return uuid.Nil, ErrMissingToken
	}

	cl := account.NewAccountsServiceClient("xmc.srv.account", client.DefaultClient)
	rsp, err := cl.Get(ctx, &account.GetRequest{ClientId: claims["sub"].(string)})
	if err != nil {
		return uuid.Nil, err
	}

	u, _ := uuid.Parse(rsp.Account.Uuid)
	return u, nil
}

func JWTWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		newCtx := ctx
		mdata, ok := metadata.FromContext(ctx)
		if ok {
			rawToken, ok := mdata["X-Jwt"]
			if ok {
				token, err := ParseToken(rawToken)
				if err == nil {
					newCtx = ContextWithToken(newCtx, token)
				}
			}
		}
		return fn(newCtx, req, rsp)
	}
}

func GetPubkey(k *api.KV) error {
	pub, err := kv.Get(k, "xmc.srv.auth/pubkey")
	if err != nil {
		return err
	}
	Pubkey, err = jwt.ParseRSAPublicKeyFromPEM(pub.Value)
	if err != nil {
		return err
	}

	return nil
}
