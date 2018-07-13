package auth

import (
	"context"
	"crypto/rsa"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/metadata"
)

func Auth(pubkey *rsa.PublicKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := request.ParseFromRequest(c.Request, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
			return pubkey, nil
		})

		hasClaims := false
		if err == nil {
			_, hasClaims = token.Claims.(jwt.MapClaims)
		}
		if err == nil && token.Valid && hasClaims {
			c.Set("JWT", metadata.NewContext(context.Background(), metadata.Metadata{"X-Jwt": token.Raw}))
			c.Set("JWTToken", token)
		} else {
			c.Set("JWT", context.Background())
		}
	}
}
