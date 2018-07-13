package problem

import (
	"github.com/google/uuid"
	pgrader "github.com/xmc-dev/xmc/xmc-core/proto/grader"
)

// Grader is a problem that gives a grade to a solution
// based on a TestCase
type Grader struct {
	ID           uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	AttachmentID uuid.UUID `gorm:"type:uuid"`
	Language     string
	Name         string `gorm:"unique_index"`
}

func GraderFromProto(gr *pgrader.Grader) *Grader {
	id, _ := uuid.Parse(gr.Id)
	attachmentID, _ := uuid.Parse(gr.AttachmentId)
	g := &Grader{
		ID:           id,
		AttachmentID: attachmentID,
		Language:     gr.Language,
		Name:         gr.Name,
	}

	return g
}

func (g *Grader) ToProto() *pgrader.Grader {
	gr := &pgrader.Grader{
		Id:           g.ID.String(),
		AttachmentId: g.AttachmentID.String(),
		Language:     g.Language,
		Name:         g.Name,
	}

	return gr
}
