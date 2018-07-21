package importer

import (
	"time"

	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
	"github.com/xmc-dev/xmc/xmc-core/proto/tsrange"
)

// TaskListSpec is a generic representation of the soon-to-be-imported Task List
type TaskListSpec struct {
	Name               string
	Description        string
	Title              string
	PublicSubmissions  bool
	WithParticipations bool
	StartTime          *time.Time
	EndTime            *time.Time

	taskListID string
}

type TaskListImporter interface {
	ReadTaskList(fp string) (*TaskListSpec, error)
}

func (tls *TaskListSpec) getTaskListID() error {
	if len(tls.taskListID) > 0 {
		return nil
	}

	log := Log.WithFields(logrus.Fields{
		"object":        "tasklist",
		"tasklist_name": tls.Name,
	})
	var err error

	tls.taskListID, err = getTaskListID(tls.Name, log)
	return err
}

func (tls *TaskListSpec) IsNew() (bool, error) {
	err := tls.getTaskListID()
	switch e := errors.Cause(err).(type) {
	case nil:
		return false, nil
	case *ErrNotFound:
		return true, nil
	default:
		return false, e
	}
}

func (tls *TaskListSpec) NeedsUpdate() (bool, error) {
	return true, nil
}

func (tls *TaskListSpec) Import() error {
	client := taskListClient()
	log := Log.WithFields(logrus.Fields{
		"object":        "tasklist",
		"tasklist_name": tls.Name,
	})

	isNew, err := tls.IsNew()
	if err != nil {
		return err
	}

	if isNew {
		log.Info("Task list is new, going to be created")
		req := &tasklist.CreateRequest{
			TaskList: &tasklist.TaskList{
				Name:               tls.Name,
				Description:        tls.Description,
				Title:              tls.Title,
				PublicSubmissions:  tls.PublicSubmissions,
				WithParticipations: tls.WithParticipations,
			},
		}
		if tls.StartTime != nil && tls.EndTime != nil {
			st, _ := ptypes.TimestampProto(*tls.StartTime)
			et, _ := ptypes.TimestampProto(*tls.EndTime)
			req.TaskList.TimeRange = &tsrange.TimestampRange{
				Begin: st,
				End:   et,
			}
		}
		_, err := client.Create(context.TODO(), req)
		if err != nil {
			return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to create task list %s", tls.Name)
		}
		log.Info("Task list successfully created")
	} else {
		needsUpdate, err := tls.NeedsUpdate()
		if err != nil {
			return err
		}
		if needsUpdate {
			req := &tasklist.UpdateRequest{
				Name:               tls.Name,
				Description:        tls.Description,
				Title:              tls.Title,
				PublicSubmissions:  &wrappers.BoolValue{Value: tls.PublicSubmissions},
				WithParticipations: &wrappers.BoolValue{Value: tls.WithParticipations},
			}
			if tls.StartTime != nil && tls.EndTime != nil {
				st, _ := ptypes.TimestampProto(*tls.StartTime)
				et, _ := ptypes.TimestampProto(*tls.EndTime)
				req.TimeRange = &tsrange.TimestampRange{
					Begin: st,
					End:   et,
				}
			}
			if err != nil {
				return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to update task list %s", tls.Name)
			}
			log.Info("Task list successfully update")
		} else {
			log.Info("Task list is up to date")
		}
	}

	return nil
}
