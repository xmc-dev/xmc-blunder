package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/account-srv/consts"
	"github.com/xmc-dev/xmc/account-srv/db"
	"github.com/xmc-dev/xmc/account-srv/proto/role"
)

type RoleService struct {
}

func roleSName(method string) string {
	return fmt.Sprintf("%s.RoleService.%s", consts.ServiceName, method)
}

func (*RoleService) Create(ctx context.Context, req *role.CreateRequest, rsp *role.CreateResponse) error {
	methodName := roleSName("Create")
	switch {
	case req.Role == nil:
		return errors.BadRequest(methodName, "missing role")
	case len(req.Role.Id) == 0:
		return errors.BadRequest(methodName, "invalid id")
	case len(req.Role.Name) == 0:
		return errors.BadRequest(methodName, "invalid name")
	case len(req.Role.Scope) == 0:
		return errors.BadRequest(methodName, "invalid scope")
	}
	req.Role.Id = strings.ToLower(req.Role.Id)

	err := db.CreateRole(req.Role)
	if err != nil {
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "id must be unique")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	return nil
}

func (*RoleService) Read(ctx context.Context, req *role.ReadRequest, rsp *role.ReadResponse) error {
	methodName := roleSName("Read")
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid id")
	}

	req.Id = strings.ToLower(req.Id)
	r, err := db.ReadRole(req.Id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "role not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	rsp.Role = r.ToProto()

	return nil
}

func (*RoleService) Update(ctx context.Context, req *role.UpdateRequest, rsp *role.UpdateResponse) error {
	methodName := roleSName("Update")
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid id")
	}

	req.Id = strings.ToLower(req.Id)
	req.NewId = strings.ToLower(req.NewId)

	err := db.UpdateRole(req)
	if err != nil {
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "id already in use")
		} else if err == db.ErrNotFound {
			return errors.NotFound(methodName, "role not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	return nil
}

func (*RoleService) Delete(ctx context.Context, req *role.DeleteRequest, rsp *role.DeleteResponse) error {
	methodName := roleSName("Delete")
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid id")
	}

	err := db.DeleteRole(strings.ToLower(req.Id))
	if err != nil {
		if e, ok := err.(db.ErrHasDependants); ok {
			return errors.BadRequest(methodName, "one or more "+string(e)+" depend on this role")
		} else if err == db.ErrNotFound {
			return errors.NotFound(methodName, "role not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	return nil
}

func (*RoleService) Search(ctx context.Context, req *role.SearchRequest, rsp *role.SearchResponse) error {
	methodName := roleSName("Search")

	if req.Limit == 0 {
		req.Limit = 10
	}

	roles, err := db.SearchRole(req)

	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}

	rsp.Roles = []*role.Role{}
	for _, r := range roles {
		rsp.Roles = append(rsp.Roles, r.ToProto())
	}

	return nil
}
