package util

import (
	"bytes"

	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/xmc-core/db"
	"github.com/xmc-dev/xmc/xmc-core/db/models/attachment"
	pattachment "github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	"github.com/xmc-dev/xmc/xmc-core/s3"
)

func MakeAttachment(d *db.Datastore, req *pattachment.CreateRequest) (uuid.UUID, error) {
	id, err := d.CreateAttachment(req.Attachment)
	if err != nil {
		return uuid.UUID{}, err
	}

	a := attachment.FromProto(req.Attachment)
	a.ID = id
	reader := bytes.NewReader(req.Contents)
	obj, err := s3.UploadAttachment(a, reader, int64(len(req.Contents)))
	if err != nil {
		return uuid.Nil, err
	}
	err = d.SetS3Object(id, obj)
	if err != nil {
		return uuid.Nil, err
	}
	err = d.SetAttachmentSize(id, int32(len(req.Contents)))
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func DeleteAttachment(d *db.Datastore, attachmentID uuid.UUID) error {
	acc, err := d.ReadAttachment(attachmentID)
	if err != nil {
		return err
	}

	err = s3.DeleteAttachment(acc)
	if err != nil {
		return err
	}
	return d.DeleteAttachment(attachmentID)
}
