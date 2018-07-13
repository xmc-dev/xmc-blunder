package importer

import (
	"context"
	"net/http"
	"os"

	"github.com/micro/go-micro/client"
	merrors "github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/registry"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	xmcconsul "github.com/xmc-dev/registry"
	"github.com/xmc-dev/xmc/xmc-core/proto/dataset"
	"github.com/xmc-dev/xmc/xmc-core/proto/grader"
	"github.com/xmc-dev/xmc/xmc-core/proto/task"
	"github.com/xmc-dev/xmc/xmc-core/proto/tasklist"
)

// Spec is the set of methods that any *Spec must implement
type Spec interface {
	// IsNew returns true if there is no such object with the Spec's name.
	// It must be implemented to prevent collisions.
	IsNew() (bool, error)

	// NeedsUpdate returns true if there is an object with the same name
	// as the Spec but with different data.
	NeedsUpdate() (bool, error)

	// Import creates or updates the internal object according to the Spec.
	Import() error
}

var Log = logrus.StandardLogger()

var BaseClient = client.NewClient(client.Registry(xmcconsul.NewRegistry(registry.Addrs(os.Getenv("MICRO_REGISTRY_ADDRESS")))))

type ErrNotFound struct {
	Object string
	Name   string
}

func (enf *ErrNotFound) Error() string {
	return "importer: " + enf.Object + " " + enf.Name + " doesn't exist"
}

func getGraderID(name string, log *logrus.Entry) (string, error) {
	log = log.WithField("grader_name", name)
	log.Info("Getting grader by name")
	rsp, err := graderClient().Get(context.TODO(), &grader.GetRequest{
		Name: name,
	})
	if err != nil {
		e := merrors.Parse(err.Error())
		if e.Code == http.StatusNotFound {
			err = &ErrNotFound{
				Object: "grader",
				Name:   name,
			}
		} else {
			err = e
		}
		return "", errors.Wrapf(err, "importer: error while getting grader", name)
	}
	log.WithField("grader_id", rsp.Grader.Id).Info("Found grader id")
	return rsp.Grader.Id, nil
}

func getDatasetID(name string, log *logrus.Entry) (string, error) {
	log = log.WithField("dataset_name", name)
	log.Info("Getting dataset")
	rsp, err := datasetClient().Get(context.TODO(), &dataset.GetRequest{
		Name: name,
	})
	if err != nil {
		e := merrors.Parse(err.Error())
		if e.Code == http.StatusNotFound {
			err = &ErrNotFound{
				Object: "dataset",
				Name:   name,
			}
		} else {
			err = e
		}
		return "", errors.Wrapf(err, "importer: error while getting dataset %s", name)
	}
	log.WithField("dataset_id", rsp.Dataset.Id).Info("Found dataset id")

	return rsp.Dataset.Id, nil
}

func getTaskID(name string, log *logrus.Entry) (string, error) {
	log = log.WithField("task_name", name)
	log.Info("Getting task")
	rsp, err := taskClient().Get(context.TODO(), &task.GetRequest{
		Name: name,
	})
	if err != nil {
		e := merrors.Parse(err.Error())
		if e.Code == http.StatusNotFound {
			err = &ErrNotFound{
				Object: "task",
				Name:   name,
			}
		} else {
			err = e
		}
		return "", errors.Wrapf(err, "importer: error while getting task %s", name)
	}
	log.WithField("task_id", rsp.Task.Id).Info("Found task id")

	return rsp.Task.Id, nil
}

func getTaskListID(name string, log *logrus.Entry) (string, error) {
	log = log.WithField("tasklist_name", name)
	log.Info("Getting task list")
	rsp, err := taskListClient().Get(context.TODO(), &tasklist.GetRequest{
		Name: name,
	})

	if err != nil {
		e := merrors.Parse(err.Error())
		if e.Code == http.StatusNotFound {
			err = &ErrNotFound{
				Object: "task",
				Name:   name,
			}
		} else {
			err = e
		}
		return "", errors.Wrapf(err, "importer: error while getting task list %s", name)
	}
	log.WithField("tasklist_id", rsp.TaskList.Id).Info("Found task list id")

	return rsp.TaskList.Id, nil
}
