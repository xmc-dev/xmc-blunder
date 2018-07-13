package storage

import (
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/auth-srv/session"
)

// Client implements the osin.Client interface
type Client struct {
	ID          string
	RedirectURI string

	Original *account.Account
}

func (c *Client) GetId() string {
	return c.ID
}

func (c *Client) GetSecret() string {
	return ""
}

func (c *Client) GetRedirectUri() string {
	return c.RedirectURI
}

func (c *Client) GetUserData() interface{} {
	return nil
}

func (c *Client) ClientSecretMatches(secret string) bool {
	rsp, err := session.AuthenticateService(c.ID, secret)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"client": c,
			"err":    err,
		}).Debug("Login request failed")
		return false
	}

	return rsp.CredentialsOk
}
