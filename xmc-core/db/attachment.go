package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/xmc-core/db/models/attachment"
	pattachment "github.com/xmc-dev/xmc/xmc-core/proto/attachment"
)

func (d *Datastore) CreateAttachment(att *pattachment.Attachment) (uuid.UUID, error) {
	a := attachment.FromProto(att)

	// because for some reason gorm refuses to do that automagically with attachments
	a.CreatedAt = time.Now()
	err := d.db.Create(a).Error
	return a.ID, e(err, "couldn't create attachment")
}

func (d *Datastore) ReadAttachment(uuid uuid.UUID) (*attachment.Attachment, error) {
	a := &attachment.Attachment{}

	err := d.db.First(a, "id = ?", uuid).Error
	return a, e(err, "couldn't read attachment")
}

func (d *Datastore) UpdateAttachment(uuid uuid.UUID, att *pattachment.UpdateRequest, s3Object string) (*attachment.Attachment, error) {
	dd := d.begin()
	a, err := dd.ReadAttachment(uuid)
	if err != nil {
		dd.Rollback()
		return nil, err
	}
	if len(att.Description) > 0 {
		a.Description = att.Description
	}
	if len(att.Filename) > 0 {
		a.Filename = att.Filename
	}
	if len(s3Object) > 0 {
		a.S3Object = s3Object
	}
	// same thing as with the creation thing in CreateAttachment
	a.UpdatedAt = time.Now()

	err = dd.db.Save(a).Error
	if err != nil {
		dd.Rollback()
		return nil, e(err, "couldn't update attachment")
	}

	return a, e(dd.Commit(), "couldn't update attachment")
}

func (d *Datastore) SetS3Object(uuid uuid.UUID, s3Object string) error {
	return e(d.db.Exec("UPDATE attachments SET s3_object=? WHERE id=?", s3Object, uuid).Error, "couldn't set attachment's S3 object")
}

func (d *Datastore) SetAttachmentSize(uuid uuid.UUID, size int32) error {
	return e(d.db.Exec("UPDATE attachments SET size = ? WHERE id = ?", size, uuid).Error, "couldn't set attachment's size")
}

func (d *Datastore) SetAttachmentPublic(uuid uuid.UUID, isPublic bool) error {
	return e(d.db.Exec("UPDATE attachments SET is_public = ? WHERE id = ?", isPublic, uuid).Error, "couldn't set attachment's public state")
}

func (d *Datastore) DeleteAttachment(uuid uuid.UUID) error {
	result := d.db.Where("id = ?", uuid).Delete(&attachment.Attachment{})
	if result.Error != nil {
		return e(result.Error, "couldn't delete attachment")
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *Datastore) SearchAttachments(req *pattachment.SearchRequest) ([]*attachment.Attachment, uint32, error) {
	dd := d.begin()
	as := []*attachment.Attachment{}
	query := dd.db
	if len(req.Description) > 0 {
		query = query.Where("description ~* ?", req.Description)
	}
	if len(req.ObjectId) > 0 {
		query = query.Where("object_id ~* ?", req.ObjectId)
	}
	if len(req.Filename) > 0 {
		query = query.Where("filename ~* ?", req.Filename)
	}
	query = query.Order("created_at DESC")
	var cnt uint32
	err := query.Model(&as).Count(&cnt).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search attachments")
	}
	query = query.Limit(req.Limit).Offset(req.Offset)
	err = query.Find(&as).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search attachments")
	}
	err = dd.Commit()
	return as, cnt, e(err, "couldn't search attachments")
}

func (d *Datastore) AttachmentExists(id uuid.UUID) (bool, error) {
	row := d.db.Raw("SELECT EXISTS(SELECT 1 FROM attachments)").Row()
	result := false

	err := row.Scan(&result)

	return result, e(err, "couldn't check for attachment existence")
}
