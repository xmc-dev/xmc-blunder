package db

import (
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/job"
	pjob "github.com/xmc-dev/xmc/dispatcher-srv/proto/job"
)

type Job interface {
	CreateJob(j *pjob.Job) (uuid.UUID, error)
	IsFinished(uuid string) bool
	ReadJob(uuid string) (*job.Job, error)
	SearchJob(req *pjob.SearchRequest) ([]*job.Job, error)
	FinishJob(req *pjob.FinishRequest) error
	SetJobStateAndEvalID(jobUUID uuid.UUID, state job.State, evalID string) error
}

func CreateJob(j *pjob.Job) (uuid.UUID, error) {
	return db.CreateJob(j)
}

func IsFinished(uuid string) bool {
	return db.IsFinished(uuid)
}

func ReadJob(uuid string) (*job.Job, error) {
	return db.ReadJob(uuid)
}

func SearchJob(req *pjob.SearchRequest) ([]*job.Job, error) {
	return db.SearchJob(req)
}

func FinishJob(req *pjob.FinishRequest) error {
	return db.FinishJob(req)
}

func SetJobStateAndEvalID(jobUUID uuid.UUID, state job.State, evalID string) error {
	return db.SetJobStateAndEvalID(jobUUID, state, evalID)
}
