package page

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
)

type ChildrenHandler struct {
	r *gin.RouterGroup
}

func (h *ChildrenHandler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.GET("/:id", h.queryEndpoint)
}

func (h *ChildrenHandler) queryEndpoint(c *gin.Context) {
	id := c.Param("id")
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	req := &page.GetFirstChildrenRequest{
		Id:     id,
		Limit:  uint32(limit),
		Offset: uint32(offset),
	}
	rsp, err := cl.GetFirstChildren(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read page's first children"))
		return
	}
	mv := []json.RawMessage{}
	for _, v := range rsp.Pages {
		mv = append(mv, util.Marshal(v))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":     meta,
		"versions": mv,
	})
}
