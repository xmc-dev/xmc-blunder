package dispatch

import (
	"github.com/micro/go-micro/client"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/dispatcher-srv/auth"
	"github.com/xmc-dev/xmc/dispatcher-srv/db"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/job"
	"github.com/xmc-dev/xmc/dispatcher-srv/status"
	"github.com/xmc-dev/xmc/eval-srv/proto/eval"
	"github.com/xmc-dev/xmc/xmc-core/proto/submission"
)

var log = logrus.WithField("prefix", "dispatch")

var submissionService = submission.NewSubmissionServiceClient("xmc.srv.core", client.DefaultClient)

func next() {
	qi, err := db.GetFirstJobInQueue()
	if err != nil {
		if err == db.ErrNotFound {
			log.Info("Nothing to dispatch")
		} else {
			handleError("couldn't get first job in queue", err)
		}
		return
	}
	log.Info("Dispatching job", qi.JobUUID)
	success := false
	nis := status.HealthCheck()
	if len(nis) == 0 {
		log.Error("No evals available")
	}
	for _, ni := range nis {
		if !ni.Idle || ni.Disabled {
			continue
		}
		err := db.SetJobStateAndEvalID(qi.JobUUID, job.PROCESSING, ni.Name)
		if err != nil {
			handleError("couldn't set job state and eval_id", err)
		} else {
			job, err := db.ReadJob(qi.JobUUID.String())
			if err != nil {
				handleError("couldn't get job", err)
			} else {
				req := client.NewRequest("xmc.srv.eval", "EvalService.Assign",
					&eval.AssignRequest{
						Job: job.ToProto(),
					})
				rsp := &eval.AssignResponse{}
				err := client.Call(auth.C(), req, rsp, client.WithAddress(ni.Address))
				if err != nil {
					handleError("assigning the job failed", err)
				} else {
					success = true
					log.WithFields(logrus.Fields{
						"qi": qi,
						"ni": ni,
					}).Info("job assigned successfully")
					_, err = submissionService.Update(auth.C(), &submission.UpdateRequest{
						Job: job.ToProto(),
					})
					if err != nil {
						handleError("updating the submission failed", err)
						success = false
					}
					break
				}
			}
		}
	}
	if !success {
		err := db.SetJobStateAndEvalID(qi.JobUUID, job.WAITING, "")
		if err != nil {
			handleError("couldn't set back job state and eval_id", err)
			return
		}
		log.WithField("qi", qi).Info("Job not dispatched")
	} else {
		_, err := db.DequeueJob()
		if err != nil {
			handleError("couldn't remove job from queue", err)
		}
	}
}

func handleError(reason string, err error) {
	log.WithError(err).Error("error while trying to dispatch job: ", reason)
}

// Next dispatches the next job in the queue
func Next() {
	go next()
}
