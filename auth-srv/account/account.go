package account

import (
	"context"

	"github.com/RangelReale/osin"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
)

var Client account.AccountsServiceClient

func GetClient(id string) (*account.Account, error) {
	rsp, err := Client.Search(context.Background(), &account.SearchRequest{
		Limit:    1,
		ClientId: id,
	})

	if err != nil {
		e := errors.Parse(err.Error())
		return nil, e
	}
	if len(rsp.Accounts) == 0 {
		return nil, osin.ErrNotFound
	}

	return rsp.Accounts[0], err
}

func ReadAccount(uuid string) (*account.ReadResponse, *errors.Error) {
	rsp, err := Client.Read(context.Background(), &account.ReadRequest{
		Uuid: uuid,
	})

	if err != nil {
		return nil, errors.Parse(err.Error())
	}

	return rsp, nil
}
