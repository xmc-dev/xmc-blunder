package handler

import (
	"context"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Handler is an API handler
type Handler interface {
	SetRouter(r *gin.RouterGroup)
}

// Register registers the API handler with the router
func Register(r *gin.RouterGroup, h Handler) {
	h.SetRouter(r)
}

// C gets the context with the JWT to be passed to server requests from the Gin context
func C(c *gin.Context) context.Context {
	ctx, ok := c.Get("JWT")
	if !ok {
		panic("JWT context not present")
	}

	return ctx.(context.Context)
}

// JWT extracts the internal JWT object from the gin context
func JWT(c *gin.Context) (*jwt.Token, bool) {
	raw, ok := c.Get("JWTToken")
	if !ok {
		return nil, ok
	}

	tok, ok := raw.(*jwt.Token)
	if !ok {
		return nil, false
	}
	return tok, ok
}
