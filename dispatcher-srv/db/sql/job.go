package sql

import (
	"time"

	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/job"
	pjob "github.com/xmc-dev/xmc/dispatcher-srv/proto/job"
)

func (s *SQL) CreateJob(j *pjob.Job) (uuid.UUID, error) {
	jb := job.FromProto(j)

	return jb.UUID, s.db.Create(jb).Error
}

func (s *SQL) IsFinished(uuuid string) bool {
	jb := job.Job{}

	err := s.db.Where("uuid = ?", uuuid).First(&jb)
	if err != nil {
		return false
	}
	return jb.State == job.DONE
}

func (s *SQL) ReadJob(uuid string) (*job.Job, error) {
	jb := job.Job{}

	err := s.db.First(&jb, "uuid = ?", uuid).Error
	if err != nil {
		return nil, e(err)
	}

	return &jb, nil
}

func (s *SQL) SearchJob(req *pjob.SearchRequest) ([]*job.Job, error) {
	where := make(map[string]interface{})
	switch {
	case len(req.TaskId) > 0:
		where["task_id"] = req.TaskId
	case len(req.DatasetId) > 0:
		where["dataset_id"] = req.DatasetId
	case len(req.Language) > 0:
		where["language"] = req.Language
	case len(req.EvalId) > 0:
		where["eval_id"] = req.EvalId
	case req.State != nil:
		where["state"] = req.State.Value
	}

	result := []*job.Job{}
	query := s.db.Where(where)
	if len(req.ErrorMessage) > 0 {
		query = query.Where("error_message ~* ?", req.ErrorMessage)
	}
	if req.Limit > 0 {
		query = query.Limit(int(req.Limit))
	}
	if req.Offset > 0 {
		query = query.Offset(int(req.Offset))
	}
	err := query.Find(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SQL) FinishJob(req *pjob.FinishRequest) error {
	j, err := s.ReadJob(req.JobUuid)
	if err != nil {
		return err
	}

	j.State = job.DONE
	t := time.Now()
	j.FinishedAt = &t
	return s.db.Save(&j).Error
}

func (s *SQL) SetJobStateAndEvalID(jobUUID uuid.UUID, state job.State, evalID string) error {
	return s.db.Exec("UPDATE jobs SET state = (CASE WHEN ? > state THEN ? ELSE state END), eval_id=? WHERE uuid=?", state, state, evalID, jobUUID).Error
}
