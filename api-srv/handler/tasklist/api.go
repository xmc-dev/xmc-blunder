package tasklist

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/micro/protobuf/ptypes"
	"github.com/pkg/errors"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
	"github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
	"github.com/xmc-dev/xmc/xmc-core/proto/tsrange"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = tasklist.NewTaskListServiceClient("xmc.srv.core", client.DefaultClient)

func getTaskList(ctx context.Context, id string) (t *tasklist.TaskList, err error) {
	_, err = uuid.Parse(id)

	if err == nil {
		req := &tasklist.ReadRequest{Id: id}
		rsp := &tasklist.ReadResponse{}
		rsp, err = cl.Read(ctx, req)
		if err == nil {
			t = rsp.TaskList
		}
	} else {
		req := &tasklist.GetRequest{Name: id}
		rsp := &tasklist.GetResponse{}
		rsp, err = cl.Get(ctx, req)
		if err == nil {
			t = rsp.TaskList
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
	h.r.GET("/:id/participate", h.participateEndpoint)
	h.r.GET("/:id/cancelparticipation", h.cancelParticipationEndpoint)
	h.r.GET("/:id/participants", h.participantsEndpoint)
}

func (h *Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &tasklist.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create task list"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	name := c.Query("name")
	description := c.Query("description")
	title := c.Query("title")
	begin, _ := time.Parse(time.RFC3339, c.Query("timeRangeBegin"))
	end, _ := time.Parse(time.RFC3339, c.Query("timeRangeEnd"))

	timeRange := &tsrange.TimestampRange{}
	if !begin.IsZero() {
		b, _ := ptypes.TimestampProto(begin)
		timeRange.Begin = b
	}
	if !end.IsZero() {
		e, _ := ptypes.TimestampProto(end)
		timeRange.End = e
	}
	if begin.IsZero() && end.IsZero() {
		timeRange = nil
	}

	req := &tasklist.SearchRequest{
		Limit:       uint32(limit),
		Offset:      uint32(offset),
		Name:        name,
		Description: description,
		TimeRange:   timeRange,
		Title:       title,
	}

	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read task lists"))
		return
	}
	ts := []json.RawMessage{}
	for _, t := range rsp.TaskLists {
		ts = append(ts, util.Marshal(t))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":      meta,
		"taskLists": ts,
	})
}

func (h *Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")
	t, err := getTaskList(handler.C(c), id)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read task list"))
		return
	}
	mt := util.Marshal(t)
	c.JSON(http.StatusOK, gin.H{
		"taskList": mt,
	})
}

func (h *Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &tasklist.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update task list"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	leaveTasks, _ := strconv.ParseBool(c.Param("leaveTasks"))
	_, err := cl.Delete(handler.C(c), &tasklist.DeleteRequest{Id: id, LeaveTasks: leaveTasks})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete task list"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (*Handler) participateEndpoint(c *gin.Context) {
	id := c.Param("id")
	_, err := cl.Participate(handler.C(c), &tasklist.ParticipateRequest{
		TaskListId: id,
	})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't participate to task list"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (*Handler) cancelParticipationEndpoint(c *gin.Context) {
	id := c.Param("id")
	_, err := cl.CancelParticipation(handler.C(c), &tasklist.CancelParticipationRequest{
		TaskListId: id,
	})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't participate to task list"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (*Handler) participantsEndpoint(c *gin.Context) {
	id := c.Param("id")
	rsp, err := cl.GetParticipants(handler.C(c), &tasklist.GetParticipantsRequest{TaskListId: id})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't get task list participants"))
		return
	}
	c.JSON(http.StatusOK, util.Marshal(rsp))
}
