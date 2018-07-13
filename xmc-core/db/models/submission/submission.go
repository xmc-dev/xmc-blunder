package submission

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	presult "github.com/xmc-dev/xmc/xmc-core/proto/result"
	psubmission "github.com/xmc-dev/xmc/xmc-core/proto/submission"
)

type State int32

const (
	// WAITING means that the job is still in the queue
	WAITING State = 0

	// PROCESSING means that the job is being processed by an eval
	PROCESSING State = 1

	// DONE means that the job has been processed
	DONE State = 2
)

type Submission struct {
	ID           uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	TaskID       uuid.UUID `gorm:"type:uuid;index"`
	DatasetID    uuid.UUID `gorm:"type:uuid"`
	AttachmentID uuid.UUID `gorm:"type:uuid"`
	UserID       uuid.UUID `gorm:"type:uuid"`
	EvalID       string
	Language     string
	CreatedAt    time.Time `gorm:"index"`
	FinishedAt   *time.Time
	State        State
}

func SubmissionFromProto(sb *psubmission.Submission) *Submission {
	s := &Submission{
		EvalID:   sb.EvalId,
		Language: sb.Language,
		State:    State(sb.State),
	}
	s.ID, _ = uuid.Parse(sb.Id)
	s.TaskID, _ = uuid.Parse(sb.TaskId)
	s.DatasetID, _ = uuid.Parse(sb.DatasetId)
	s.AttachmentID, _ = uuid.Parse(sb.AttachmentId)
	s.UserID, _ = uuid.Parse(sb.UserId)
	s.CreatedAt, _ = ptypes.Timestamp(sb.CreatedAt)
	finishedAt, _ := ptypes.Timestamp(sb.FinishedAt)
	s.FinishedAt = &finishedAt

	return s
}

func (s *Submission) ToProto(r *presult.Result) *psubmission.Submission {
	sb := &psubmission.Submission{
		Id:           s.ID.String(),
		TaskId:       s.TaskID.String(),
		DatasetId:    s.DatasetID.String(),
		AttachmentId: s.AttachmentID.String(),
		UserId:       s.UserID.String(),
		EvalId:       s.EvalID,
		Language:     s.Language,
		State:        psubmission.State(s.State),
		Result:       r,
	}
	sb.CreatedAt, _ = ptypes.TimestampProto(s.CreatedAt)
	if s.FinishedAt != nil {
		sb.FinishedAt, _ = ptypes.TimestampProto(*s.FinishedAt)
	}

	return sb
}

type Result struct {
	SubmissionID       uuid.UUID `gorm:"type:uuid;primary_key"`
	ErrorMessage       string
	CompilationMessage string
	Score              decimal.Decimal
	BuildCommand       string
}

func (r *Result) ToProto(tr []*presult.TestResult) *presult.Result {
	rs := &presult.Result{
		ErrorMessage:       r.ErrorMessage,
		CompilationMessage: r.CompilationMessage,
		Score:              r.Score.String(),
		TestResults:        tr,
		BuildCommand:       r.BuildCommand,
	}

	return rs
}

func (Result) TableName() string {
	return "submission_results"
}

type TestResult struct {
	SubmissionID  uuid.UUID `gorm:"type:uuid;primary_key"`
	TestNo        int32     `gorm:"primary_key"`
	Score         decimal.Decimal
	GraderMessage string
	Memory        int32
	Time          time.Duration
}

func (t *TestResult) ToProto() *presult.TestResult {
	tr := &presult.TestResult{
		TestNo:        t.TestNo,
		Score:         t.Score.String(),
		GraderMessage: t.GraderMessage,
		Memory:        t.Memory,
		Time:          ptypes.DurationProto(t.Time),
	}

	return tr
}
