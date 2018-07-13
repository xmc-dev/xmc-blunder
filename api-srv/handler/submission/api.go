package submission

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/micro/protobuf/jsonpb"
	"github.com/micro/protobuf/ptypes"
	"github.com/pkg/errors"
	e "github.com/xmc-dev/xmc/api-srv/errors"
	"github.com/xmc-dev/xmc/api-srv/handler"
	"github.com/xmc-dev/xmc/api-srv/util"
	"github.com/xmc-dev/xmc/xmc-core/proto/submission"
	"github.com/xmc-dev/xmc/xmc-core/proto/tsrange"
)

// Handler is the submission API handler
type Handler struct {
	r *gin.RouterGroup
}

var cl = submission.NewSubmissionServiceClient("xmc.srv.core", client.DefaultClient)
var mrsh = &jsonpb.Marshaler{}

func (h *Handler) SetRouter(r *gin.RouterGroup) {
	h.r = r
	h.r.GET("/", h.queryEndpoint)
	h.r.POST("/", h.createEndpoint)
	h.r.GET("/:id", h.readEndpoint)
	h.r.DELETE("/:id", h.deleteEndpoint)
}

func (h *Handler) queryEndpoint(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("perPage"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	taskID := c.Query("taskId")
	datasetID := c.Query("datasetId")
	userID := c.Query("userId")
	evalID := c.Query("evalId")
	state := c.Query("state")
	language := c.Query("language")
	createdAtBegin, _ := time.Parse(time.RFC3339, c.Query("createdAtBegin"))
	createdAtEnd, _ := time.Parse(time.RFC3339, c.Query("createdAtEnd"))
	finishedAtBegin, _ := time.Parse(time.RFC3339, c.Query("finishedAtBegin"))
	finishedAtEnd, _ := time.Parse(time.RFC3339, c.Query("finishedAtEnd"))
	errorMessage := c.Query("ErrorMessage")
	compilationMessage := c.Query("compilationMessage")
	includeResult, _ := strconv.ParseBool(c.Query("includeResult"))
	includeTestResults, _ := strconv.ParseBool(c.Query("includeTestResults"))

	var createdAt, finishedAt *tsrange.TimestampRange
	var stateValue *submission.StateValue

	if len(state) != 0 {
		val, ok := submission.State_value[strings.ToUpper(state)]
		if ok {
			stateValue = &submission.StateValue{
				Value: submission.State(val),
			}
		}
	}
	if !createdAtBegin.IsZero() || !createdAtEnd.IsZero() {
		beginP, _ := ptypes.TimestampProto(createdAtBegin)
		endP, _ := ptypes.TimestampProto(createdAtEnd)
		createdAt = &tsrange.TimestampRange{Begin: beginP, End: endP}
	}
	if !finishedAtBegin.IsZero() || !finishedAtEnd.IsZero() {
		beginP, _ := ptypes.TimestampProto(finishedAtBegin)
		endP, _ := ptypes.TimestampProto(finishedAtEnd)
		finishedAt = &tsrange.TimestampRange{Begin: beginP, End: endP}
	}

	req := &submission.SearchRequest{
		Limit:              uint32(limit),
		Offset:             uint32(offset),
		TaskId:             taskID,
		DatasetId:          datasetID,
		UserId:             userID,
		EvalId:             evalID,
		State:              stateValue,
		Language:           language,
		CreatedAt:          createdAt,
		FinishedAt:         finishedAt,
		ErrorMessage:       errorMessage,
		CompilationMessage: compilationMessage,
		IncludeResult:      includeResult,
		IncludeTestResults: includeTestResults,
	}
	rsp, err := cl.Search(handler.C(c), req)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read submissions"))
		return
	}
	ms := []json.RawMessage{}
	for _, s := range rsp.Submissions {
		ms = append(ms, util.Marshal(s))
	}
	meta := util.Marshal(rsp.Meta)
	c.JSON(http.StatusOK, gin.H{
		"meta":        meta,
		"submissions": ms,
	})
}

func (h *Handler) createEndpoint(c *gin.Context) {
	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		e.BadRequest(c)
		return
	}
	toCreate := &submission.CreateRequest{}
	err = util.Unmarshal(string(buf), toCreate)
	if err != nil {
		e.BadRequest(c)
		return
	}
	rsp, err := cl.Create(handler.C(c), toCreate)
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(err, "couldn't create submission"))
		return
	}

	r := util.Marshal(rsp)
	c.JSON(http.StatusOK, r)
}

func (h *Handler) readEndpoint(c *gin.Context) {
	id := c.Param("id")
	ir, _ := strconv.ParseBool(c.Query("includeResult"))
	itr, _ := strconv.ParseBool(c.Query("includeTestResults"))
	rsp, err := cl.Read(handler.C(c), &submission.ReadRequest{Id: id, IncludeResult: ir, IncludeTestResults: itr})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't read submissions"))
		return
	}
	m := util.Marshal(rsp.Submission)
	c.JSON(http.StatusOK, gin.H{"submission": m})
}

func (h *Handler) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	_, err := cl.Delete(handler.C(c), &submission.DeleteRequest{Id: id})
	if err != nil {
		me := merrors.Parse(err.Error())
		e.Error(c, me, errors.Wrap(me, "couldn't delete submissions"))
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
