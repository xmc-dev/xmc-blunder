package handler

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/xmc-core/common"
	"github.com/xmc-dev/xmc/xmc-core/db"
	"github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	"github.com/xmc-dev/xmc/xmc-core/proto/grader"
	"github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"
	"github.com/xmc-dev/xmc/xmc-core/s3"
	"github.com/xmc-dev/xmc/xmc-core/util"
)

type GraderService struct{}

func graderSName(method string) string {
	return fmt.Sprintf("%s.GraderService.%s", "xmc.srv.core", method)
}

func (*GraderService) Create(ctx context.Context, req *grader.CreateRequest, rsp *grader.CreateResponse) error {
	methodName := graderSName("Create")
	switch {
	case req.Grader == nil:
		return errors.BadRequest(methodName, "missing grader")
	case req.Code == nil:
		return errors.BadRequest(methodName, "invalid code")
	case !common.IsValidLanguage(req.Grader.Language):
		return errors.BadRequest(methodName, "invalid language")
	case len(req.Grader.Name) == 0:
		return errors.BadRequest(methodName, "invalid name")
	}

	req.Grader.AttachmentId = ""
	req.Grader.Id = ""

	dd := db.DB.BeginGroup()
	id, err := dd.CreateGrader(req.Grader)
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	attID, err := util.MakeAttachment(dd, &attachment.CreateRequest{
		Attachment: &attachment.Attachment{
			ObjectId: "grader/" + id.String(),
			Filename: "grader." + req.Grader.Language,
		},
		Contents: req.Code,
	})
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.GraderSetAttachmentID(id, attID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Id = id.String()
	return nil
}

func (*GraderService) Read(ctx context.Context, req *grader.ReadRequest, rsp *grader.ReadResponse) error {
	methodName := graderSName("Read")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	g, err := db.DB.ReadGrader(id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "grader not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Grader = g.ToProto()

	return nil
}

func (*GraderService) Get(ctx context.Context, req *grader.GetRequest, rsp *grader.GetResponse) error {
	methodName := graderSName("Get")

	if len(req.Name) == 0 {
		return errors.BadRequest(methodName, "invalid name")
	}

	g, err := db.DB.GetGrader(req.Name)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "grader not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Grader = g.ToProto()

	return nil
}

func (*GraderService) Update(ctx context.Context, req *grader.UpdateRequest, rsp *grader.UpdateResponse) error {
	methodName := graderSName("Update")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	if req.Code != nil {
		g, err := dd.ReadGrader(id)
		if err != nil {
			dd.Rollback()
			if err == db.ErrNotFound {
				return errors.NotFound(methodName, "grader not found")
			}
			return errors.InternalServerError(methodName, e(err))
		}
		att, err := dd.ReadAttachment(g.AttachmentID)
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}

		reader := bytes.NewReader(req.Code)
		_, err = s3.UploadAttachment(att, reader, int64(len(req.Code)))
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
	}

	err = dd.UpdateGrader(req)
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*GraderService) Delete(ctx context.Context, req *grader.DeleteRequest, rsp *grader.DeleteResponse) error {
	methodName := graderSName("Delete")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	g, err := dd.ReadGrader(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "grader not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.DeleteGrader(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "grader not found")
		} else if e, ok := err.(db.ErrHasDependants); ok {
			return errors.BadRequest(methodName, "one or more "+string(e)+" depend on this grader")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	err = util.DeleteAttachment(dd, g.AttachmentID)
	if err != nil {
		dd.Rollback()
		log.Warnf("Error while deleting grader attachment id %v: %v", g.AttachmentID, err)
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil

}

func (*GraderService) Search(ctx context.Context, req *grader.SearchRequest, rsp *grader.SearchResponse) error {
	methodName := graderSName("Search")

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	gs, total, err := db.DB.SearchGrader(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	grs := []*grader.Grader{}
	for _, g := range gs {
		grs = append(grs, g.ToProto())
	}

	rsp.Graders = grs
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(grs)),
		Total:   total,
	}

	return nil
}
