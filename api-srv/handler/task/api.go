package task

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
	"github.com/xmc-dev/xmc/xmc-core/proto/task"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = task.NewTaskServiceClient("xmc.srv.core", client.DefaultClient)

func getTask(ctx context.Context, id string) (t *task.Task, err error) {
	_, err = uuid.Parse(id)

	if err == nil {
		req := &task.ReadRequest{Id: id}
		rsp := &task.ReadResponse{}
		rsp, err = cl.Read(ctx, req)
		if err == nil {
			t = rsp.Task
		}
	} else {
		req := &task.GetRequest{Name: id}
		rsp := &task.GetResponse{}
		rsp, err = cl.Get(ctx, req)
		if err == nil {
			t = rsp.Task
		}
	}

	return
}

func (h *Handler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.POST("/", h.createEndpoint)
	h.r.GET("/", h.queryEndpoint)
	h.r.GET("/:id", h.readEndpoint)
	h.r.PATCH("/:id", h.updateEndpoint)
	h.r.DELETE("/:id", h.deleteEndpoint)
}

func (h *Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &task.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create task"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	datasetID := c.Query("datasetId")
	name := c.Query("name")
	description := c.Query("description")
	title := c.Query("title")
	taskListID := c.Query("taskListId")

	req := &task.SearchRequest{
		Limit:       uint32(limit),
		Offset:      uint32(offset),
		DatasetId:   datasetID,
		Name:        name,
		Description: description,
		Title:       title,
		TaskListId:  taskListID,
	}
	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read tasks"))
		return
	}
	mt := []json.RawMessage{}
	for _, t := range rsp.Tasks {
		mt = append(mt, util.Marshal(t))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":  meta,
		"tasks": mt,
	})
}

func (h *Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")
	t, err := getTask(handler.C(c), id)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read task"))
		return
	}
	mt := util.Marshal(t)
	c.JSON(http.StatusOK, gin.H{
		"task": mt,
	})
}

func (h *Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &task.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update task"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	_, err := cl.Delete(handler.C(c), &task.DeleteRequest{Id: id})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete task"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
