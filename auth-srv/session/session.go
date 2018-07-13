package session

import (
	"context"

	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/account-srv/proto/session"
)

var Client session.SessionsServiceClient

func ReadSession(sessionUUID string) (*session.Session, *errors.Error) {
	rsp, err := Client.Read(context.Background(), &session.ReadRequest{
		Uuid: sessionUUID,
	})

	if err != nil {
		return nil, errors.Parse(err.Error())
	}
	return rsp.Session, nil
}

func Login(username, password string) (*session.LoginResponse, *errors.Error) {
	rsp, err := Client.Login(context.Background(), &session.LoginRequest{
		ClientId:     username,
		ClientSecret: password,
		Type:         &account.TypeValue{Value: account.Type_USER},
	})

	if err != nil {
		return nil, errors.Parse(err.Error())
	}

	return rsp, nil
}

func Logout(sessionUUID string) (*session.LogoutResponse, *errors.Error) {
	rsp, err := Client.Logout(context.Background(), &session.LogoutRequest{
		Uuid: sessionUUID,
	})

	if err != nil {
		return nil, errors.Parse(err.Error())
	}
	return rsp, nil
}

func AuthenticateService(clientID, clientSecret string) (*session.LoginResponse, *errors.Error) {
	rsp, err := Client.Login(context.Background(), &session.LoginRequest{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		Type:         &account.TypeValue{Value: account.Type_SERVICE},
	})

	if err != nil {
		return nil, errors.Parse(err.Error())
	}

	return rsp, nil
}
