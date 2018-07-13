// Package req provides utility functions for making authenticated requests between XMC components.
package req

import (
	"context"

	"github.com/micro/go-micro/metadata"
	"golang.org/x/oauth2"
)

// C returns a context that can be fed to a micro-generated request function
func C(src oauth2.TokenSource, meta ...metadata.Metadata) (context.Context, error) {
	tok, err := src.Token()
	if err != nil {
		return nil, err
	}
	m := metadata.Metadata{
		"X-Jwt": tok.AccessToken,
	}
	for _, me := range meta {
		for k, v := range me {
			m[k] = v
		}
	}

	return metadata.NewContext(context.Background(), m), nil
}
