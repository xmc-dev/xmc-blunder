package db

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xmc-dev/xmc/xmc-core/db/models/submission"
	psubmission "github.com/xmc-dev/xmc/xmc-core/proto/submission"
)

func (d *Datastore) CreateSubmission(sb *psubmission.Submission) (uuid.UUID, error) {
	ss := submission.SubmissionFromProto(sb)

	err := d.db.Create(ss).Error
	return ss.ID, e(err, "couldn't create submission")
}

func (d *Datastore) ReadSubmission(id uuid.UUID) (*submission.Submission, error) {
	ss := &submission.Submission{}

	err := d.db.Where("id = ?", id).First(ss).Error

	return ss, e(err, "couldn't read submission")
}

func (d *Datastore) ReadSubmissionResult(id uuid.UUID) (*submission.Result, error) {
	sr := &submission.Result{}

	err := d.db.Where("submission_id = ?", id).First(sr).Error

	return sr, e(err, "couldn't read submission result")
}

func (d *Datastore) ReadTestResults(id uuid.UUID) ([]*submission.TestResult, error) {
	tr := []*submission.TestResult{}

	err := d.db.Where("submission_id = ?", id).Find(&tr).Error

	return tr, e(err, "couldn't read test results")
}

func (d *Datastore) UpdateSubmission(req *psubmission.UpdateRequest) error {
	dd := d.begin()
	id, _ := uuid.Parse(req.Job.SubmissionId)
	var ss *submission.Submission
	var sr submission.Result
	var err error
	ss, err = dd.ReadSubmission(id)
	if err != nil {
		return err
	}
	ss.EvalID = req.Job.EvalId
	ss.State = submission.State(req.Job.State)
	finishedAt, _ := ptypes.Timestamp(req.Job.FinishedAt)
	ss.FinishedAt = &finishedAt
	err = dd.db.Save(ss).Error
	if err != nil {
		dd.Rollback()
		return err
	}

	if req.Job.Result != nil {
		err = dd.db.FirstOrCreate(&sr, submission.Result{SubmissionID: id}).Error
		if err != nil {
			dd.Rollback()
			return e(err, "couldn't create or save submission result")
		}
		sr.ErrorMessage = req.Job.Result.ErrorMessage
		sr.CompilationMessage = req.Job.Result.CompilationMessage
		sr.Score, _ = decimal.NewFromString(req.Job.Result.Score)
		sr.BuildCommand = req.Job.Result.BuildCommand
		err = dd.db.Save(&sr).Error
		if err != nil {
			dd.Rollback()
			return e(err, "couldn't save submission result")
		}
		for _, pt := range req.Job.Result.TestResults {
			t := submission.TestResult{}
			err = dd.db.FirstOrCreate(&t, submission.TestResult{
				SubmissionID: id,
				TestNo:       pt.TestNo,
			}).Error
			if err != nil {
				dd.Rollback()
				return e(err, "couldn't create or save test result")
			}
			t.Score, _ = decimal.NewFromString(pt.Score)
			t.GraderMessage = pt.GraderMessage
			t.Memory = pt.Memory
			t.Time, _ = ptypes.Duration(pt.Time)
			err = dd.db.Save(&t).Error
			if err != nil {
				dd.Rollback()
				return e(err, "couldn't save test result")
			}
		}
	}

	return e(dd.Commit(), "couldn't update submission")
}

func (d *Datastore) DeleteSubmission(id uuid.UUID) error {
	result := d.db.Where("id = ?", id).Delete(&submission.Submission{})
	if result.Error != nil {
		return e(result.Error, "couldn't delete submission")
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *Datastore) SearchSubmission(req *psubmission.SearchRequest) ([]*submission.Submission, uint32, error) {
	dd := d.begin()
	ss := []*submission.Submission{}
	query := dd.db.Joins("FULL OUTER JOIN submission_results on submissions.id = submission_results.submission_id")

	if len(req.TaskId) > 0 {
		query = query.Where("submissions.task_id = ?", req.TaskId)
	}
	if len(req.DatasetId) > 0 {
		query = query.Where("submissions.dataset_id = ?", req.DatasetId)
	}
	if len(req.EvalId) > 0 {
		query = query.Where("submissions.eval_id = ?", req.EvalId)
	}
	if req.State != nil {
		query = query.Where("submissions.state = ?", req.State.Value)
	}
	if len(req.Language) > 0 {
		query = query.Where("submissions.Language = ?", req.Language)
	}
	if req.CreatedAt != nil {
		createdAtBegin, _ := ptypes.Timestamp(req.CreatedAt.Begin)
		createdAtEnd, _ := ptypes.Timestamp(req.CreatedAt.End)
		r := tsrange(createdAtBegin, createdAtEnd)
		query = query.Where("submissions.created_at <@ ?::tstzrange", r)
	}
	if req.FinishedAt != nil {
		finishedAtBegin, _ := ptypes.Timestamp(req.FinishedAt.Begin)
		finishedAtEnd, _ := ptypes.Timestamp(req.FinishedAt.End)
		r := tsrange(finishedAtBegin, finishedAtEnd)
		query = query.Where("submissions.finished_at <@ ?::tstzrange", r)
	}
	if len(req.UserId) > 0 {
		u, _ := uuid.Parse(req.UserId)
		query = query.Where("submissions.user_id = ?", u)
	}

	// result fields
	if len(req.ErrorMessage) > 0 {
		query = query.Where("submission_results.error_message ~* ?", req.ErrorMessage)
	}
	if len(req.CompilationMessage) > 0 {
		query = query.Where("submission_results.compilation_message ~* ?", req.CompilationMessage)
	}
	var cnt uint32
	err := query.Model(&ss).Count(&cnt).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search submissions")
	}
	query = query.Limit(req.Limit).Offset(req.Offset).Order("created_at DESC, finished_at DESC")
	err = query.Find(&ss).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search submissions")
	}
	err = dd.Commit()
	return ss, cnt, e(err, "couldn't search submissions")
}

func (d *Datastore) SubmissionSetAttachmentID(id, attID uuid.UUID) error {
	return e(d.db.Exec("UPDATE submissions SET attachment_id = ? WHERE id = ?", attID, id).Error, "couldn't set submission's attachment id")
}

func (d *Datastore) SubmissionSetDatasetID(id, datasetID uuid.UUID) error {
	return e(d.db.Exec("UPDATE submissions SET dataset_id = ? WHERE id = ?", datasetID, id).Error, "couldn't set submission's dataset id")
}
