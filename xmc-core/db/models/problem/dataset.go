package problem

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	pdataset "github.com/xmc-dev/xmc/xmc-core/proto/dataset"
)

// Dataset stores the information necessary for the evaluation of a submission,
// like the grader's code, tests etc.
type Dataset struct {
	ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	Name        string    `gorm:"unique_index"`
	GraderID    uuid.UUID `gorm:"type:uuid"`
	Description string
	TimeLimit   time.Duration
	MemoryLimit int32
}

func DatasetFromProto(ds *pdataset.Dataset) *Dataset {
	id, _ := uuid.Parse(ds.Id)
	graderID, _ := uuid.Parse(ds.GraderId)
	d := &Dataset{
		ID:          id,
		Name:        ds.Name,
		GraderID:    graderID,
		Description: ds.Description,
		MemoryLimit: ds.MemoryLimit,
	}
	d.TimeLimit, _ = ptypes.Duration(ds.TimeLimit)

	return d
}

func (d *Dataset) ToProto() *pdataset.Dataset {
	ds := &pdataset.Dataset{
		Id:          d.ID.String(),
		Name:        d.Name,
		GraderId:    d.GraderID.String(),
		Description: d.Description,
		TimeLimit:   ptypes.DurationProto(d.TimeLimit),
		MemoryLimit: d.MemoryLimit,
	}

	return ds
}
