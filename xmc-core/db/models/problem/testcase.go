package problem

import (
	"github.com/google/uuid"
	pdataset "github.com/xmc-dev/xmc/xmc-core/proto/dataset"
)

// TestCase is a test of the dataset
type TestCase struct {
	ID                 uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	DatasetID          uuid.UUID `gorm:"type:uuid;unique_index:idx_dataset_id_number"`
	Number             int32     `gorm:"unique_index:idx_dataset_id_number"`
	InputAttachmentID  uuid.UUID `gorm:"type:uuid"`
	OutputAttachmentID uuid.UUID `gorm:"type:uuid"`
}

func (tc *TestCase) ToProto() *pdataset.TestCase {
	return &pdataset.TestCase{
		Id:                 tc.ID.String(),
		Number:             tc.Number,
		InputAttachmentId:  tc.InputAttachmentID.String(),
		OutputAttachmentId: tc.OutputAttachmentID.String(),
	}
}
