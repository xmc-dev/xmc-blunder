package attachment

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
	"github.com/xmc-dev/xmc/xmc-core/proto/attachment"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = attachment.NewAttachmentServiceClient("xmc.srv.core", client.DefaultClient)

func (h *Handler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.POST("/", h.createEndpoint)
	h.r.GET("/", h.queryEndpoint)
	h.r.GET("/:id", h.readEndpoint)
	h.r.GET("/:id/download", h.downloadEndpoint)
	h.r.GET("/:id/file", h.fileEndpoint)
	h.r.PATCH("/:id", h.updateEndpoint)
	h.r.DELETE("/:id", h.deleteEndpoint)
}

func (h *Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &attachment.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create attachment"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	objectClass := c.Query("objectClass")
	objectID := c.Query("objectId")
	filename := c.Query("filename")
	desc := c.Query("description")
	oid := ""

	// if objectClass was given, get attachments in that class
	// i.e. their objectId starts with the objectClass
	// if the objectID was given too then only query for exact matches
	if len(objectClass) > 0 {
		oid += "^" + objectClass + "/"
		if len(objectID) > 0 {
			oid += objectID + "$"
		}
	}
	req := &attachment.SearchRequest{
		Limit:       uint32(limit),
		Offset:      uint32(offset),
		ObjectId:    oid,
		Filename:    filename,
		Description: desc,
	}
	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read attachments"))
		return
	}
	ma := []json.RawMessage{}
	for _, a := range rsp.Attachments {
		ma = append(ma, util.Marshal(a))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":        meta,
		"attachments": ma,
	})
}

func (h *Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")
	req := &attachment.ReadRequest{Id: id}
	rsp, err := cl.Read(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read attachment"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"attachment": util.Marshal(rsp.Attachment),
	})
}

func (h *Handler) downloadEndpoint(c *gin.Context) {
	id := c.Param("id")
	req := &attachment.GetContentsRequest{Id: id}
	rsp, err := cl.GetContents(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't get attachment url"))
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, rsp.Url)
}

func (h *Handler) fileEndpoint(c *gin.Context) {
	id := c.Param("id")
	req := &attachment.GetContentsRequest{Id: id}
	rsp, err := cl.GetContents(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't get attachment url"))
		return
	}
	c.JSON(http.StatusOK, util.Marshal(rsp))
}

func (h *Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	public := c.Query("public")

	setPublic, err := strconv.ParseBool(public)
	if err == nil {
		_, err = cl.SetPublic(handler.C(c), &attachment.SetPublicRequest{
			AttachmentId: id,
			Public:       setPublic,
		})
		if err != nil {
			me := merrors.Parse(err.Error())
			e.Error(c, me, errors.Wrap(err, "couldn't set public status of attachment"))
			return
		}
	} else if len(public) > 0 {
		e.BadRequest(c)
		return
	}

	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &attachment.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		if err == io.EOF {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update attachment"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	req := &attachment.DeleteRequest{Id: id}
	_, err := cl.Delete(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete attachment"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
