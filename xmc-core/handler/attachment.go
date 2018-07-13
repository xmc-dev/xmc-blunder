package handler

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/xmc-core/db"
	mattachment "github.com/xmc-dev/xmc/xmc-core/db/models/attachment"
	"github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	"github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"
	"github.com/xmc-dev/xmc/xmc-core/s3"
	"github.com/xmc-dev/xmc/xmc-core/util"
)

type AttachmentService struct {
}

func attachmentSName(method string) string {
	return fmt.Sprintf("%s.AttachmentService.%s", "xmc.srv.core", method)
}

func (*AttachmentService) Create(ctx context.Context, req *attachment.CreateRequest, rsp *attachment.CreateResponse) error {
	methodName := attachmentSName("Create")
	if !perms.HasScope(ctx, "manage/attachment") {
		return errors.Forbidden(methodName, "you are not allowed to create attachments")
	}
	switch {
	case req.Attachment == nil:
		return errors.BadRequest(methodName, "missing attachment")
	case len(req.Contents) == 0:
		return errors.BadRequest(methodName, "invalid contents")
	case len(req.Attachment.Filename) == 0:
		return errors.BadRequest(methodName, "invalid filename")
	case len(req.Attachment.ObjectId) == 0:
		return errors.BadRequest(methodName, "invalid object_id")
	}

	dd := db.DB.BeginGroup()
	id, err := util.MakeAttachment(dd, req)
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "object_id and filename must be unique in pair")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Id = id.String()
	return nil
}

func (*AttachmentService) Read(ctx context.Context, req *attachment.ReadRequest, rsp *attachment.ReadResponse) error {
	methodName := attachmentSName("Read")
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	a, err := db.DB.ReadAttachment(uuid)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "attachment not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	if !a.IsPublic && !perms.HasScope(ctx, "manage/attachment") {
		return errors.Forbidden(methodName, "you are not allowed to read this attachment")
	}
	rsp.Attachment = a.ToProto()
	return nil
}

func (*AttachmentService) GetContents(ctx context.Context, req *attachment.GetContentsRequest, rsp *attachment.GetContentsResponse) error {
	methodName := attachmentSName("GetContents")
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	att, err := db.DB.ReadAttachment(uuid)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "attachment not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if !att.IsPublic && !perms.HasScope(ctx, "manage/attachment") {
		return errors.Forbidden(methodName, "you are not allowed to get the contents of this attachment")
	}

	url, err := s3.GetURL(att, "")
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Url = url.String()
	return nil
}

func (*AttachmentService) SetPublic(ctx context.Context, req *attachment.SetPublicRequest, rsp *attachment.SetPublicResponse) error {
	methodName := attachmentSName("SetPublic")
	if !perms.HasScope(ctx, "manage/attachment") {
		return errors.Forbidden(methodName, "you are not allowed to manage this attachment")
	}

	if len(req.AttachmentId) == 0 {
		return errors.BadRequest(methodName, "invalid attachment_id")
	}

	var att *mattachment.Attachment
	uuid, err := uuid.Parse(req.AttachmentId)
	if err != nil {
		return errors.BadRequest(methodName, "invalid attachment_id")
	}

	dd := db.DB.BeginGroup()
	att, err = dd.ReadAttachment(uuid)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "attachment not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.SetAttachmentPublic(att.ID, req.Public)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*AttachmentService) Update(ctx context.Context, req *attachment.UpdateRequest, rsp *attachment.UpdateResponse) error {
	methodName := attachmentSName("Update")
	if !perms.HasScope(ctx, "manage/attachment") {
		return errors.Forbidden(methodName, "you are not allowed to update this attachment")
	}
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid id")
	}

	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	att, err := dd.ReadAttachment(uuid)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "attachment not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	// if the filename changed then we need to update the s3 object
	s3Object := ""
	if len(req.Filename) > 0 {
		s3Object, err = s3.RenameAttachment(att, req.Filename)
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
		err = dd.SetS3Object(uuid, s3Object)
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
	}

	att, err = dd.UpdateAttachment(uuid, req, s3Object)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "attachment not found")
		} else if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "object_id and filename must be unique in pair")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if req.Contents != nil {
		reader := bytes.NewReader(req.Contents)
		_, err := s3.UploadAttachment(att, reader, int64(len(req.Contents)))
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
		err = dd.SetAttachmentSize(uuid, int32(len(req.Contents)))
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*AttachmentService) Delete(ctx context.Context, req *attachment.DeleteRequest, rsp *attachment.DeleteResponse) error {
	methodName := attachmentSName("Delete")
	if !perms.HasScope(ctx, "manage/attachment") {
		return errors.Forbidden(methodName, "you are not allowed to delete this attachment")
	}
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	uuid, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	dd := db.DB.BeginGroup()
	err = util.DeleteAttachment(dd, uuid)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "attachment not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*AttachmentService) Search(ctx context.Context, req *attachment.SearchRequest, rsp *attachment.SearchResponse) error {
	methodName := attachmentSName("Search")

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	atts, total, err := db.DB.SearchAttachments(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	as := []*attachment.Attachment{}
	for _, att := range atts {
		as = append(as, att.ToProto())
	}

	rsp.Attachments = as
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(as)),
		Total:   total,
	}
	return nil
}
