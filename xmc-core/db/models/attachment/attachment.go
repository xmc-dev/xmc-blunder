package attachment

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	pattachment "github.com/xmc-dev/xmc/xmc-core/proto/attachment"
)

// Attachment is used to store additional files.
type Attachment struct {
	ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	S3Object    string
	Description string
	ObjectID    string `gorm:"unique_index:idx_object_id_filename"`
	Filename    string `gorm:"unique_index:idx_object_id_filename"`
	Size        int32
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func FromProto(att *pattachment.Attachment) *Attachment {
	uuid, _ := uuid.Parse(att.Id)
	a := &Attachment{
		ID:          uuid,
		S3Object:    att.S3Object,
		Description: att.Description,
		ObjectID:    att.ObjectId,
		Filename:    att.Filename,
		Size:        att.Size,
		IsPublic:    att.IsPublic,
	}
	a.CreatedAt, _ = ptypes.Timestamp(att.CreatedAt)
	a.UpdatedAt, _ = ptypes.Timestamp(att.UpdatedAt)

	return a
}

func (a *Attachment) ToProto() *pattachment.Attachment {
	att := &pattachment.Attachment{
		Id:          a.ID.String(),
		S3Object:    a.S3Object,
		Description: a.Description,
		ObjectId:    a.ObjectID,
		Filename:    a.Filename,
		Size:        a.Size,
		IsPublic:    a.IsPublic,
	}
	att.CreatedAt, _ = ptypes.TimestampProto(a.CreatedAt)
	att.UpdatedAt, _ = ptypes.TimestampProto(a.UpdatedAt)

	return att
}
