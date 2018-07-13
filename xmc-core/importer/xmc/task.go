package xmc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/xmc-dev/xmc/xmc-core/importer"
	yaml "gopkg.in/yaml.v2"
)

// TaskImporter imports tasks stored in the XMC format.
//
// In the XMC format, the task is a directory with the name of the task
// that contains a single file named task.yaml that contains the details of the task.
//
// The task.yaml file has the following structure:
//
//	description: Some task description
//	dataset_name: example_dataset
//	input_file: problem.in
//	output_file: problem.out
//  task_list_name: archive
//
// The task.yaml must be a valid YAML file. input_file and output_file can also have
// the stdin and stdout value respectively.
type TaskImporter struct {
}

type internalTaskSpec struct {
	Description  string `yaml:"description"`
	DatasetName  string `yaml:"dataset_name"`
	InputFile    string `yaml:"input_file"`
	OutputFile   string `yaml:"output_file"`
	TaskListName string `yaml:"task_list_name"`
}

func NewTaskImporter() *TaskImporter {
	return &TaskImporter{}
}

func (ti *TaskImporter) ReadTask(fp string) (*importer.TaskSpec, error) {
	fi, err := os.Stat(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-task-importer: couldn't stat file %s", fp)
	}
	if !fi.IsDir() {
		return nil, errors.New("xmc-task-importer: invalid task, path " + fp + " is not a directory")
	}

	ts := &importer.TaskSpec{}
	ts.Name = strings.ToLower(filepath.Base(fp))

	specFile, err := ioutil.ReadFile(filepath.Join(fp, "task.yaml"))
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-task-importer: failed to read task.yaml")
	}

	is := internalTaskSpec{}
	err = yaml.Unmarshal(specFile, &is)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-task-importer: error in parsing task.yaml")
	}

	ts.Description = is.Description
	ts.DatasetName = is.DatasetName
	ts.InputFile = is.InputFile
	ts.OutputFile = is.OutputFile
	ts.TaskListName = is.TaskListName

	return ts, nil
}
