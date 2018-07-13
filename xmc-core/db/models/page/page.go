package page

import (
	"time"

	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	ppage "github.com/xmc-dev/xmc/xmc-core/proto/page"
)

// Page represents a wiki page
type Page struct {
	ID              uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	Path            string    `gorm:"unique_index;type:ltree"`
	LatestTimestamp time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
}

func FromProto(pg *ppage.Page) *Page {
	id, _ := uuid.Parse(pg.Id)
	p := &Page{
		ID:   id,
		Path: pg.Path,
	}
	p.LatestTimestamp, _ = ptypes.Timestamp(pg.LatestTimestamp)

	return p
}

func (p *Page) ToProto() *ppage.Page {
	pg := &ppage.Page{
		Id:   p.ID.String(),
		Path: "/" + strings.Replace(p.Path, ".", "/", -1),
	}
	pg.LatestTimestamp, _ = ptypes.TimestampProto(p.LatestTimestamp)

	return pg
}

// Version represents a revision of a Page
type Version struct {
	PageID       uuid.UUID `gorm:"primary_key;type:uuid"`
	AttachmentID uuid.UUID `gorm:"type:uuid"`
	Title        string
	Timestamp    time.Time `gorm:"primary_key"`
	DeletedAt    *time.Time
}

func (Version) TableName() string {
	return "page_versions"
}

func (v *Version) ToProto() *ppage.Version {
	ver := &ppage.Version{
		PageId:       v.PageID.String(),
		AttachmentId: v.AttachmentID.String(),
		Title:        v.Title,
	}
	ver.Timestamp, _ = ptypes.TimestampProto(v.Timestamp)

	return ver
}
