package job

import (
	"time"

	"github.com/google/uuid"
	"github.com/micro/protobuf/ptypes"
	"github.com/xmc-dev/xmc/dispatcher-srv/proto/job"
)

// Code holds source code
type Code []byte

func (c Code) String() string {
	return string([]byte(c))
}

type State int32

const (
	// WAITING means that the job is still in the queue
	WAITING State = 0

	// PROCESSING means that the job is being processed by an eval
	PROCESSING State = 1

	// DONE means that the job has been processed
	DONE State = 2
)

// Job is an evaluation job
type Job struct {
	UUID         uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v1mc()"`
	DatasetID    string
	Code         []byte
	Language     string
	EvalID       string
	State        State
	Deviation    int32
	SubmissionID string
	TaskID       string
	CreatedAt    time.Time
	FinishedAt   *time.Time
}

func FromProto(j *job.Job) *Job {
	jb := &Job{}
	var err error

	jb.UUID, _ = uuid.Parse(j.Uuid)
	jb.DatasetID = j.DatasetId
	jb.Code = j.Code
	jb.Language = j.Language
	jb.EvalID = j.EvalId
	// extra safe-guard
	jb.State = State(j.State)
	jb.SubmissionID = j.SubmissionId
	jb.TaskID = j.TaskId
	jb.CreatedAt, err = ptypes.Timestamp(j.CreatedAt)
	if err != nil {
		panic(err)
	}
	finishedAt, err := ptypes.Timestamp(j.FinishedAt)
	jb.FinishedAt = &finishedAt
	if err != nil {
		jb.FinishedAt = nil
	}

	return jb
}

func (j *Job) ToProto() *job.Job {
	var err error
	pj := job.Job{
		Uuid:         j.UUID.String(),
		DatasetId:    j.DatasetID,
		Code:         j.Code,
		Language:     j.Language,
		EvalId:       j.EvalID,
		State:        job.State(j.State),
		SubmissionId: j.SubmissionID,
		TaskId:       j.TaskID,
	}
	pj.CreatedAt, err = ptypes.TimestampProto(j.CreatedAt)
	if err != nil {
		panic(err)
	}
	if j.FinishedAt != nil {
		pj.FinishedAt, err = ptypes.TimestampProto(*j.FinishedAt)
		if err != nil {
			panic(err)
		}
	}

	return &pj
}
