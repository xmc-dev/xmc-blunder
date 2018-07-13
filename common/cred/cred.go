// Package cred provides utility functions for getting and using OAuth2 credentials for XMC components.
package cred

import (
	"golang.org/x/oauth2"
	"github.com/hashicorp/consul/api"
	"github.com/xmc-dev/xmc/common/kv"
	"strings"
	"errors"
	"golang.org/x/oauth2/clientcredentials"
	"context"
)

var Src oauth2.TokenSource

func InitAuth(serviceName string, k *api.KV, tokenURL string, scopes ...string) error {
	cred, err := kv.Get(k, serviceName+"/credentials")
	if err != nil {
		return err
	}

	parts := strings.Split(string(cred.Value), ":")
	if len(parts) != 2 {
		return errors.New("invalid credentials")
	}

	conf := clientcredentials.Config{
		ClientID: parts[0],
		ClientSecret: parts[1],
		TokenURL: tokenURL,
		Scopes: scopes,
	}
	Src = conf.TokenSource(context.Background())
	tok, err := Src.Token()
	if err != nil {
		return err
	}
	Src = oauth2.ReuseTokenSource(tok, Src)

	return nil
}
