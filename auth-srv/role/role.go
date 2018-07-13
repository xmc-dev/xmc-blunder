package role

import (
	"context"

	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/account-srv/proto/role"
)

var Client role.RoleServiceClient

func ReadRole(id string) (*role.Role, *errors.Error) {
	req := &role.ReadRequest{
		Id: id,
	}
	rsp, err := Client.Read(context.Background(), req)
	if err != nil {
		return nil, errors.Parse(err.Error())
	}
	return rsp.Role, nil
}
