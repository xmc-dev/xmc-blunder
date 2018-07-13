package account

import (
	"context"
	"encoding/json"
	"net/http"

	"io/ioutil"

	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
	"github.com/xmc-dev/xmc/api-srv/handler/role"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = account.NewAccountsServiceClient("xmc.srv.account", client.DefaultClient)

func getUser(ctx context.Context, id string) (a *account.Account, err error) {
	_, err = uuid.Parse(id)

	if err == nil {
		req := &account.ReadRequest{Uuid: id}
		rsp := &account.ReadResponse{}
		rsp, err = cl.Read(ctx, req)
		if err == nil {
			a = rsp.Account
		}
	} else {
		req := &account.GetRequest{ClientId: id}
		rsp := &account.GetResponse{}
		rsp, err = cl.Get(ctx, req)
		if err == nil {
			a = rsp.Account
		}
	}

	return
}


func getUserInfo(ctx context.Context, id string) (map[string]json.RawMessage, error) {
	a, err := getUser(ctx, id)
	if err != nil {
		return nil, err
	}
	rsp := map[string]json.RawMessage{
		"account": util.Marshal(a),
	}
	if a.Type == account.Type_USER {
		r, err := role.GetRole(ctx, a.RoleId)
		if err != nil {
			return nil, err
		}
		rsp["role"] = util.Marshal(r)
	}

	return rsp, nil
}

func (h *Handler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.POST("/", h.createEndpoint)
	h.r.GET("/:id", h.readEndpoint)
	h.r.GET("/", h.queryEndpoint)
	h.r.PATCH("/:id", h.updateEndpoint)
	h.r.DELETE("/:id", h.deleteEndpoint)
}

func (*Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &account.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create account"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (*Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")
	if id == "me" {
		tok, ok := handler.JWT(c)
		if !ok {
			e.Unauthorized(c)
			return
		}
		id, ok = tok.Claims.(jwt.MapClaims)["sub"].(string)
		if !ok {
			e.InternalServerError(c)
			return
		}
	}
	ma, err := getUserInfo(handler.C(c), id)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read account"))
		return
	}
	c.JSON(http.StatusOK, ma)
}

func (*Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	clientID := c.Query("clientId")
	rawType := c.Query("type")
	ownerUUID := c.Query("ownerUuid")
	callbackURL := c.Query("callbackUrl")
	name := c.Query("name")
	rawIsPublic := c.Query("isPublic")
	roleID := c.Query("roleId")

	req := &account.SearchRequest{
		Limit:       uint32(limit),
		Offset:      uint32(offset),
		ClientId:    clientID,
		OwnerUuid:   ownerUUID,
		CallbackUrl: callbackURL,
		Name:        name,
		RoleId:      roleID,
	}

	if ty, ok := account.Type_value[strings.ToUpper(rawType)]; ok {
		req.Type = &account.TypeValue{Value: account.Type(ty)}
	}
	if isPublic, err := strconv.ParseBool(rawIsPublic); err == nil {
		req.IsPublic = &wrappers.BoolValue{Value: isPublic}
	}

	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read accounts"))
		return
	}
	ma := []json.RawMessage{}
	for _, a := range rsp.Accounts {
		ma = append(ma, util.Marshal(a))
	}
	c.JSON(http.StatusOK, gin.H{
		"accounts": ma,
	})
}

func (*Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &account.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Uuid = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update account"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	_, err := cl.Delete(handler.C(c), &account.DeleteRequest{Uuid: id})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete account"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
