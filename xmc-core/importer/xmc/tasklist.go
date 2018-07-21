package xmc

import (
	"os"
	"path/filepath"
	"time"

	"strings"

	"io/ioutil"

	"fmt"

	"github.com/pkg/errors"
	"github.com/xmc-dev/xmc/xmc-core/importer"
	"gopkg.in/yaml.v2"
)

// TaskListImporter imports task lists stored in the XMC format.
//
// In the XMC format, the task list is a directory with the name of the task list
// that contains a single file named tasklist.yaml that containts the details of the
// task list.
//
// The tasklist.yaml has the following structure:
//
//	description: Some tasklist description
//	title: Title of its page
//	public_submissions: false
//	with_participations: false
//	start_time: 2018-07-20T11:25:40+02:00 # optional
//  end_time: 2018-07-20T13:25:40+02:00 # required only if start_time is present
//
// The tasklist.yaml must be a valid YAML file.
type TaskListImporter struct {
}

type internalTaskListSpec struct {
	Description        string    `yaml:"description"`
	Title              string    `yaml:"title"`
	PublicSubmissions  bool      `yaml:"public_submissions"`
	WithParticipations bool      `yaml:"with_participations"`
	StartTime          time.Time `yaml:"start_time"`
	EndTime            time.Time `yaml:"end_time"`
}

func NewTaskListImporter() *TaskListImporter {
	return &TaskListImporter{}
}

func (tli *TaskListImporter) ReadTaskList(fp string) (*importer.TaskListSpec, error) {
	fi, err := os.Stat(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-task-list-importer: couldn't stat file %s", fp)
	}
	if !fi.IsDir() {
		return nil, errors.New("xmc-task-list-importer: invlid task list, path " + fp + "is not a directory")
	}

	tls := &importer.TaskListSpec{}
	tls.Name = strings.ToLower(filepath.Base(fp))

	specFile, err := ioutil.ReadFile(filepath.Join(fp, "tasklist.yaml"))
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-task-list-importer: failes to read tasklist.yaml")
	}

	is := internalTaskListSpec{}
	err = yaml.Unmarshal(specFile, &is)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-task-list-importer: error in parsing tasklist.yaml")
	}
	fmt.Println("!!is ", is)

	tls.Description = is.Description
	tls.Title = is.Title
	tls.PublicSubmissions = is.PublicSubmissions
	tls.WithParticipations = is.WithParticipations
	if !is.StartTime.IsZero() {
		if is.EndTime.IsZero() {
			return nil, errors.New("xmc-task-list-importer: start_time present without an end_time")
		}
		tls.StartTime = &is.StartTime
		tls.EndTime = &is.EndTime
	}
	fmt.Println("!!tls ", tls)

	return tls, nil
}
