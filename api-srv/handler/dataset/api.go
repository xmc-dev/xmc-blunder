package dataset

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
	"github.com/xmc-dev/xmc/xmc-core/proto/dataset"
)

type Handler struct {
	r *gin.RouterGroup
}

var cl = dataset.NewDatasetServiceClient("xmc.srv.core", client.DefaultClient)

func getDataset(ctx context.Context, id string) (d *dataset.Dataset, err error) {
	_, err = uuid.Parse(id)

	if err == nil {
		req := &dataset.ReadRequest{Id: id}
		rsp := &dataset.ReadResponse{}
		rsp, err = cl.Read(ctx, req)
		if err == nil {
			d = rsp.Dataset
		}
	} else {
		req := &dataset.GetRequest{Name: id}
		rsp := &dataset.GetResponse{}
		rsp, err = cl.Get(ctx, req)
		if err == nil {
			d = rsp.Dataset
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

	tcr := h.r.Group("/:id/testcases")
	tcr.POST("/:nr", h.addTestCaseEndpoint)
	tcr.GET("/", h.queryTestCasesEndpoint)
	tcr.GET("/:nr", h.readTestCaseEndpoint)
	tcr.PATCH("/:nr", h.updateTestCaseEndpoint)
	tcr.DELETE("/:nr", h.deleteTestCaseEndpoint)
	tcr.GET("/:nr/:file", h.downloadTestCaseEndpoint)
}

func (h *Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &dataset.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create dataset"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	graderID := c.Query("graderId")
	description := c.Query("description")
	name := c.Query("name")

	req := &dataset.SearchRequest{
		Limit:       uint32(limit),
		Offset:      uint32(offset),
		GraderId:    graderID,
		Description: description,
		Name:        name,
	}
	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read datasets"))
		return
	}
	md := []json.RawMessage{}
	for _, d := range rsp.Datasets {
		md = append(md, util.Marshal(d))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":     meta,
		"datasets": md,
	})
}

func (h *Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")

	d, err := getDataset(handler.C(c), id)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read dataset"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dataset": util.Marshal(d),
	})
}

func (h *Handler) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &dataset.UpdateRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	rsp, err := cl.Update(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update dataset"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	_, err := cl.Delete(handler.C(c), &dataset.DeleteRequest{Id: id})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete dataset"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) queryTestCasesEndpoint(c *gin.Context) {
	id := c.Param("id")
	req := &dataset.GetTestCasesRequest{Id: id}
	rsp, err := cl.GetTestCases(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read test cases"))
		return
	}
	mtc := []json.RawMessage{}
	for _, tc := range rsp.TestCases {
		mtc = append(mtc, util.Marshal(tc))
	}
	c.JSON(http.StatusOK, gin.H{"testCases": mtc})
}

func (h *Handler) addTestCaseEndpoint(c *gin.Context) {
	id := c.Param("id")
	nr := c.Param("nr")
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &dataset.AddTestCaseRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate.Id = id
	intNr, err := strconv.ParseInt(nr, 10, 32)
	if err != nil {
		e.NotFound(c)
		return
	}
	toCreate.Number = int32(intNr)
	rsp, err := cl.AddTestCase(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create dataset"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) readTestCaseEndpoint(c *gin.Context) {
	id := c.Param("id")
	nr, err := strconv.ParseInt(c.Param("nr"), 10, 32)
	if err != nil {
		e.NotFound(c)
		return
	}

	req := &dataset.GetTestCaseRequest{
		Id:     id,
		Number: int32(nr),
	}
	rsp, err := cl.GetTestCase(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read test case"))
		return
	}

	c.JSON(http.StatusOK, util.Marshal(rsp.TestCase))
}

func (h *Handler) updateTestCaseEndpoint(c *gin.Context) {
	id := c.Param("id")
	nr, err := strconv.ParseInt(c.Param("nr"), 10, 32)
	if err != nil {
		e.NotFound(c)
		return
	}

	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate := &dataset.UpdateTestCaseRequest{}
	err = util.Unmarshal(string(buf), toUpdate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toUpdate.Id = id
	toUpdate.Number = int32(nr)
	rsp, err := cl.UpdateTestCase(handler.C(c), toUpdate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't update test case"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) deleteTestCaseEndpoint(c *gin.Context) {
	id := c.Param("id")
	nr, err := strconv.ParseInt(c.Param("nr"), 10, 32)
	if err != nil {
		e.NotFound(c)
		return
	}

	req := &dataset.RemoveTestCaseRequest{
		Id:     id,
		Number: int32(nr),
	}
	_, err = cl.RemoveTestCase(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete test case"))
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) downloadTestCaseEndpoint(c *gin.Context) {
	id := c.Param("id")
	nr, err := strconv.ParseInt(c.Param("nr"), 10, 32)
	file := c.Param("file")
	if err != nil || (file != "in" && file != "ok") {
		e.NotFound(c)
		return
	}

	req := &dataset.GetTestCaseRequest{
		Id:     id,
		Number: int32(nr),
	}
	rsp, err := cl.GetTestCase(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read test case"))
		return
	}

	att := ""
	if file == "in" {
		att = rsp.TestCase.InputAttachmentId
	} else {
		att = rsp.TestCase.OutputAttachmentId
	}

	fwd := c.Request.Header.Get("X-Forwarded-Prefix")
	c.Redirect(http.StatusMovedPermanently, fwd+"/attachments/"+att+"/file")
}
