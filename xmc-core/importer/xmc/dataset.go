package xmc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/xmc-dev/xmc/xmc-core/importer"
	yaml "gopkg.in/yaml.v2"
)

// DatasetImporter imports datasets stored in the XMC format.
//
// In the XMC format, the dataset is a directory with the name of the dataset
// in which there is a dataset.yaml file that contains dataset metadata, and a directory named testcases
// which contains the test files of the dataset.
//
// The testcases directory holds for each test case two files:
// an input file, for example test1.in, and an output file, test1.out.
// The format of the filename is test#.in/test#.ok, where # is the number of the test case.
// The case numbering must start from 1 and they all must be consecutive.
//
// The dataset.yaml file has the following structure:
//
//	description: Some dataset description
//	grader_name: example_grader
//	memory_limit: 1024
//	time_limit: 1.23s
//
// The dataset.yaml file must be a valid YAML file. The memory limit is expressed in bytes
// and the time limit is in the format accepted by Go's library function time.ParseDuration.
// In short, it is a sequence of integers, each with an optional fraction and a unit suffix.
// Unit suffixes are "ns", "us", "ms", "s", "m", "h".
type DatasetImporter struct {
}

type internalDatasetSpec struct {
	GraderName  string `yaml:"grader_name"`
	Description string `yaml:"description"`
	MemoryLimit int32  `yaml:"memory_limit"`
	TimeLimit   string `yaml:"time_limit"`
}

func NewDatasetImporter() *DatasetImporter {
	return &DatasetImporter{}
}

var rx = regexp.MustCompile(`^test([0-9]+)\.(in|ok)$`)

func (di *DatasetImporter) readTestCases(path string) ([]*importer.TestCaseSpec, error) {
	fp := filepath.Join(path, "testcases")
	files, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, errors.Wrap(err, "xmc-dataset-importer: couldn't read testcases directory")
	}

	input := make(map[int][]byte)
	output := make(map[int][]byte)
	maxIn := 0
	maxOut := 0
	for _, f := range files {
		match := rx.FindStringSubmatch(f.Name())
		if match == nil {
			return nil, errors.New("xmc-dataset-importer: test case file " + f.Name() + " doesn't match format")
		}
		ext := match[2]
		no, err := strconv.Atoi(match[1])
		if err != nil {
			panic(err)
		}
		if ext == "in" {
			if no > maxIn {
				maxIn = no
			}
			in, err := ioutil.ReadFile(filepath.Join(fp, f.Name()))
			if err != nil {
				return nil, errors.Wrap(err, "xmc-dataset-importer: couldn't read input file "+f.Name())
			}
			input[no] = in
		} else {
			if no > maxOut {
				maxOut = no
			}
			out, err := ioutil.ReadFile(filepath.Join(fp, f.Name()))
			if err != nil {
				return nil, errors.Wrap(err, "xmc-dataset-importer: couldn't read output file "+f.Name())
			}
			output[no] = out
		}
	}

	if maxIn != maxOut {
		return nil, errors.New("xmc-dataset-importer: number of input files doesn't equal number of output files")
	}
	for i := 1; i <= maxIn; i++ {
		_, ok := input[i]
		if !ok {
			return nil, errors.New("xmc-dataset-importer: no input file for test #" + strconv.Itoa(i))
		}
		_, ok = output[i]
		if !ok {
			return nil, errors.New("xmc-dataset-importer: no output file for test #" + strconv.Itoa(i))
		}
	}

	tcs := []*importer.TestCaseSpec{}
	for i := 1; i <= maxIn; i++ {
		tcs = append(tcs, &importer.TestCaseSpec{
			Number: int32(i),
			Input:  input[i],
			Output: output[i],
		})
	}

	return tcs, nil
}

func (di *DatasetImporter) ReadDataset(fp string) (*importer.DatasetSpec, error) {
	fi, err := os.Stat(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-dataset-importer: couldn't stat file %s", fp)
	}
	if !fi.IsDir() {
		return nil, errors.New("xmc-dataset-importer: invalid dataset, path " + fp + " is not a directory")
	}

	ds := &importer.DatasetSpec{}
	ds.Name = strings.ToLower(filepath.Base(fp))

	specFile, err := ioutil.ReadFile(filepath.Join(fp, "dataset.yaml"))
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-dataset-importer: failed to read dataset.yaml")
	}

	is := internalDatasetSpec{}
	err = yaml.Unmarshal(specFile, &is)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-dataset-importer: error in parsing dataset.yaml")
	}

	ds.Description = is.Description
	ds.GraderName = is.GraderName
	ds.MemoryLimit = is.MemoryLimit
	ds.TimeLimit, err = time.ParseDuration(is.TimeLimit)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-dataset-importer: couldn't parse time limit '%s'", is.TimeLimit)
	}

	ds.TestCases, err = di.readTestCases(fp)

	return ds, err
}
