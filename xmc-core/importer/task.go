package importer

import (
	"context"

	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/xmc-core/proto/task"
	"github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
)

// TaskSpec is a generic represenatation of the soon-to-be-imported Task
type TaskSpec struct {
	Name         string
	DatasetName  string
	TaskListName string
	Description  string
	InputFile    string
	OutputFile   string

	datasetID  string
	taskID     string
	taskListID string
}

type TaskImporter interface {
	ReadTask(fp string) (*TaskSpec, error)
}

func taskClient() task.TaskServiceClient {
	return task.NewTaskServiceClient("xmc.srv.core", BaseClient)
}

func taskListClient() tasklist.TaskListServiceClient {
	return tasklist.NewTaskListServiceClient("xmc.srv.core", client.DefaultClient)
}

func (ts *TaskSpec) getDatasetID() error {
	if len(ts.datasetID) > 0 {
		return nil
	}

	log := Log.WithFields(logrus.Fields{
		"object":       "task",
		"dataset_name": ts.DatasetName,
	})
	var err error

	ts.datasetID, err = getDatasetID(ts.DatasetName, log)

	return err
}

func (ts *TaskSpec) getTaskID() error {
	if len(ts.taskID) > 0 {
		return nil
	}
	log := Log.WithFields(logrus.Fields{
		"object":    "task",
		"task_name": ts.Name,
	})
	var err error

	ts.taskID, err = getTaskID(ts.Name, log)

	return err
}

func (ts *TaskSpec) getTaskListID() error {
	if len(ts.taskListID) > 0 {
		return nil
	}

	log := Log.WithFields(logrus.Fields{
		"object":        "tasklist",
		"tasklist_name": ts.TaskListName,
	})
	var err error

	ts.taskListID, err = getTaskListID(ts.TaskListName, log)
	return err
}

func (ts *TaskSpec) IsNew() (bool, error) {
	err := ts.getTaskID()
	switch e := errors.Cause(err).(type) {
	case nil:
		return false, nil
	case *ErrNotFound:
		return true, nil
	default:
		return false, e
	}
}

func (ts *TaskSpec) NeedsUpdate() (bool, error) {
	return true, nil
}

func (ts *TaskSpec) Import() error {
	client := taskClient()
	log := Log.WithFields(logrus.Fields{
		"object":    "task",
		"task_name": ts.Name,
	})

	isNew, err := ts.IsNew()
	if err != nil {
		return err
	}

	if isNew {
		log.Info("Task is new, going to be created")
		err := ts.getDatasetID()
		if err != nil {
			return err
		}
		err = ts.getTaskListID()
		if err != nil {
			return err
		}
		_, err = client.Create(context.TODO(), &task.CreateRequest{
			Task: &task.Task{
				Name:        ts.Name,
				DatasetId:   ts.datasetID,
				Description: ts.Description,
				TaskListId:  ts.taskListID,
				InputFile:   ts.InputFile,
				OutputFile:  ts.OutputFile,
			},
		})
		if err != nil {
			return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to create task %s", ts.Name)
		}
		log.Info("Task successfully created")
	} else {
		needsUpdate, err := ts.NeedsUpdate()
		if err != nil {
			return err
		}
		if needsUpdate {
			err = ts.getDatasetID()
			if err != nil {
				return err
			}
			err = ts.getTaskID()
			if err != nil {
				return err
			}
			err = ts.getTaskListID()
			if err != nil {
				return err
			}
			_, err = client.Update(context.TODO(), &task.UpdateRequest{
				Id:          ts.taskID,
				Description: ts.Description,
				Name:        ts.Name,
				DatasetId:   ts.datasetID,
				TaskListId:  ts.taskListID,
				InputFile:   ts.InputFile,
				OutputFile:  ts.OutputFile,
			})
			if err != nil {
				return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to update task %s", ts.Name)
			}
			log.Info("Task successfully updated")
		} else {
			log.Info("Task is up to date")
		}
	}

	return nil
}
