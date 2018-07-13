package page

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/micro/protobuf/ptypes"
	"github.com/pkg/errors"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = page.NewPageServiceClient("xmc.srv.core", client.DefaultClient)

func getPage(ctx context.Context, id string, rest string, tstamp *time.Time) (p *page.Page, err error) {
	_, err = uuid.Parse(id)

	if err == nil {
		var pt *timestamp.Timestamp
		if tstamp != nil {
			pt, _ = ptypes.TimestampProto(*tstamp)
		}
		req := &page.ReadRequest{Id: id, Timestamp: pt}
		rsp := &page.ReadResponse{}
		rsp, err = cl.Read(ctx, req)
		if err == nil {
			p = rsp.Page
		}
	} else {
		req := &page.GetRequest{Path: "/" + id + rest}
		rsp := &page.GetResponse{}
		rsp, err = cl.Get(ctx, req)
		if err == nil {
			p = rsp.Page
		}
	}

	return
}

func (h *Handler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.POST("/", h.createEndpoint)
	h.r.GET("/:id", h.readEndpoint)
	h.r.GET("/:id/*rest", h.readEndpoint)
	h.r.GET("/", h.queryEndpoint)
	h.r.PATCH("/:id", h.updateEndpoint)
	h.r.DELETE("/:id", h.deleteEndpoint)
}

func (h *Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &page.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create page"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")
	timestamp := c.Query("timestamp")
	var t *time.Time
	if len(timestamp) > 0 {
		x, _ := time.Parse(time.RFC3339, timestamp)
		t = &x
	}
	if id == "<root>" {
		id = ""
	}
	rest := c.Param("rest")
	p, err := getPage(handler.C(c), id, rest, t)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read page"))
		return
	}
	mp := util.Marshal(p)
	c.JSON(http.StatusOK, gin.H{
		"page": mp,
	})
}

func (h *Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	path := c.Query("path")
	title := c.Query("title")

	req := &page.SearchRequest{
		Limit:  uint32(limit),
		Offset: uint32(offset),
		Path:   path,
		Title: title,
	}
	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read pages"))
		return
	}
	mp := []json.RawMessage{}
	for _, p := range rsp.Pages {
		mp = append(mp, util.Marshal(p))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":  meta,
		"pages": mp,
	})
}

func (h *Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &page.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update page"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	hard, _ := strconv.ParseBool(c.Query("hard"))
	_, err := cl.Delete(handler.C(c), &page.DeleteRequest{Id: id, Hard: hard})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete page"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
