package grader

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
	"github.com/xmc-dev/xmc/xmc-core/proto/grader"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = grader.NewGraderServiceClient("xmc.srv.core", client.DefaultClient)

func getGrader(ctx context.Context, id string) (g *grader.Grader, err error) {
	_, err = uuid.Parse(id)

	if err == nil {
		req := &grader.ReadRequest{Id: id}
		rsp := &grader.ReadResponse{}
		rsp, err = cl.Read(ctx, req)
		if err == nil {
			g = rsp.Grader
		}
	} else {
		req := &grader.GetRequest{Name: id}
		rsp := &grader.GetResponse{}
		rsp, err = cl.Get(ctx, req)
		if err == nil {
			g = rsp.Grader
		}
	}

	return
}

func (h *Handler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.POST("/", h.createEndpoint)
	h.r.GET("/:id", h.readEndpoint)
	h.r.GET("/", h.queryEndpoint)
	h.r.PATCH("/:id", h.updateEndpoint)
	h.r.DELETE("/:id", h.deleteEndpoint)
	h.r.GET("/:id/file", h.fileEndpoint)
}

func (h *Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &grader.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create grader"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")

	g, err := getGrader(handler.C(c), id)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read grader"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"grader": util.Marshal(g),
	})
}

func (h *Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	name := c.Query("name")
	language := c.Query("language")

	req := &grader.SearchRequest{
		Limit:    uint32(limit),
		Offset:   uint32(offset),
		Name:     name,
		Language: language,
	}
	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read graders"))
		return
	}
	mg := []json.RawMessage{}
	for _, g := range rsp.Graders {
		mg = append(mg, util.Marshal(g))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":    meta,
		"graders": mg,
	})
}

func (h *Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &grader.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't update grader"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")

	req := &grader.DeleteRequest{Id: id}
	_, err := cl.Delete(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete grader"))
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) fileEndpoint(c *gin.Context) {
	id := c.Param("id")
	req := &grader.ReadRequest{Id: id}
	rsp, err := cl.Read(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read grader"))
		return
	}
	fwd := c.Request.Header.Get("X-Forwarded-Prefix")
	c.Redirect(http.StatusMovedPermanently, fwd+"/attachments/"+rsp.Grader.AttachmentId+"/file")
}
