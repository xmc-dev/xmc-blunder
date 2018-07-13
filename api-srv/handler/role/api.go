package role

import (
	"context"
	"strconv"

	"encoding/json"

	"net/http"

	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	"github.com/xmc-dev/xmc/account-srv/proto/role"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = role.NewRoleServiceClient("xmc.srv.account", client.DefaultClient)

func GetRole(ctx context.Context, id string) (*role.Role, error) {
	rsp, err := cl.Read(ctx, &role.ReadRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return rsp.Role, nil
}

func (h *Handler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.POST("/", h.createEndpoint)
	h.r.GET("/", h.queryEndpoint)
	h.r.GET("/:id", h.readEndpoint)
	h.r.PATCH("/:id", h.updateEndpoint)
	h.r.DELETE("/:id", h.deleteEndpoint)
}

func (*Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &role.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create role"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (*Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	id := c.Query("id")
	name := c.Query("name")
	scope := c.Query("scope")

	req := &role.SearchRequest{
		Limit:  uint32(limit),
		Offset: uint32(offset),
		Id:     id,
		Name:   name,
		Scope:  scope,
	}

	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read roles"))
		return
	}
	mr := []json.RawMessage{}
	for _, r := range rsp.Roles {
		mr = append(mr, util.Marshal(r))
	}
	c.JSON(http.StatusOK, gin.H{
		"roles": mr,
	})
}

func (*Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")
	r, err := GetRole(handler.C(c), id)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read role"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role": util.Marshal(r),
	})
}

func (*Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &role.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update role"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (*Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	_, err := cl.Delete(handler.C(c), &role.DeleteRequest{Id: id})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete role"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
