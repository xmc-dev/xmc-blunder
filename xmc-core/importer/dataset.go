package importer

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/xmc-core/proto/dataset"
)

// DatasetSpec is a generic representation of the soon-to-be-imported Dataset
type DatasetSpec struct {
	Name        string
	GraderName  string
	Description string
	MemoryLimit int32
	TimeLimit   time.Duration
	TestCases   []*TestCaseSpec

	graderID  string
	datasetID string
}

// TestCaseSpec is a generic representation of the soon-to-be-imported TestCase of the DatasetSpec
type TestCaseSpec struct {
	DatasetID string
	Number    int32
	Input     []byte
	Output    []byte
}

type DatasetImporter interface {
	ReadDataset(fp string) (*DatasetSpec, error)
}

func datasetClient() dataset.DatasetServiceClient {
	return dataset.NewDatasetServiceClient("xmc.srv.core", BaseClient)
}

func (ds *DatasetSpec) getGraderID() error {
	if len(ds.graderID) > 0 {
		return nil
	}

	log := Log.WithFields(logrus.Fields{
		"object":       "dataset",
		"dataset_name": ds.Name,
	})
	var err error

	ds.graderID, err = getGraderID(ds.GraderName, log)

	return err
}

func (ds *DatasetSpec) getDatasetID() error {
	if len(ds.datasetID) > 0 {
		return nil
	}

	log := Log.WithFields(logrus.Fields{
		"object":       "dataset",
		"dataset_name": ds.Name,
	})
	var err error

	ds.datasetID, err = getDatasetID(ds.Name, log)

	return err
}

func (ds *DatasetSpec) IsNew() (bool, error) {
	err := ds.getDatasetID()
	switch e := errors.Cause(err).(type) {
	case nil:
		return false, nil
	case *ErrNotFound:
		return true, nil
	default:
		return false, e
	}
}

func (ds *DatasetSpec) NeedsUpdate() (bool, error) {
	return true, nil
}

func (tcs *TestCaseSpec) IsNew() (bool, error) {
	client := datasetClient()
	rsp, err := client.GetTestCases(context.TODO(), &dataset.GetTestCasesRequest{
		Id: tcs.DatasetID,
	})

	if err != nil {
		e := merrors.Parse(err.Error())
		return false, errors.Wrapf(e, "importer: checking for a new test case #%d for dataset %s failed", tcs.Number, tcs.DatasetID)
	}

	exists := false
	for _, tc := range rsp.TestCases {
		if tc.Number == tcs.Number {
			exists = true
			break
		}
	}

	return !exists, nil
}

func (tcs *TestCaseSpec) NeedsUpdate() (bool, error) {
	return true, nil
}

func (ds *DatasetSpec) Import() error {
	client := datasetClient()
	log := Log.WithFields(logrus.Fields{
		"object":       "dataset",
		"dataset_name": ds.Name,
	})

	isNew, err := ds.IsNew()
	if err != nil {
		return err
	}

	if isNew {
		log.Info("Dataset is new, going to be created")
		err := ds.getGraderID()
		if err != nil {
			return err
		}
		_, err = client.Create(context.TODO(), &dataset.CreateRequest{
			Dataset: &dataset.Dataset{
				Name:        ds.Name,
				GraderId:    ds.graderID,
				Description: ds.Description,
				MemoryLimit: ds.MemoryLimit,
				TimeLimit:   ptypes.DurationProto(ds.TimeLimit),
			},
		})
		if err != nil {
			return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to create dataset %s", ds.Name)
		}
		log.Info("Dataset successfully created")
	} else {
		needsUpdate, err := ds.NeedsUpdate()
		if err != nil {
			return err
		}
		if needsUpdate {
			log.Info("Dataset needs update, going to be updated")
			err = ds.getGraderID()
			if err != nil {
				return err
			}
			err = ds.getDatasetID()
			if err != nil {
				return err
			}
			_, err = client.Update(context.TODO(), &dataset.UpdateRequest{
				Id:          ds.datasetID,
				Description: ds.Description,
				GraderId:    ds.graderID,
			})
			if err != nil {
				return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to update dataset %s", ds.Name)
			}
			log.Info("Dataset successfully updated")
		} else {
			log.Info("Dataset is up to date")
		}
	}

	log.Info("Importing test cases")
	err = ds.getDatasetID()
	if err != nil {
		return err
	}
	for _, tc := range ds.TestCases {
		tc.DatasetID = ds.datasetID
		err := tc.Import()
		if err != nil {
			return err
		}
	}

	return nil
}

func (tcs *TestCaseSpec) Import() error {
	client := datasetClient()
	log := Log.WithFields(logrus.Fields{
		"object":      "testcase",
		"testcase_no": tcs.Number,
		"dataset_id":  tcs.DatasetID,
	})

	isNew, err := tcs.IsNew()
	if err != nil {
		return err
	}

	if isNew {
		log.Info("Test case is new, going to be added to the dataset")
		_, err = client.AddTestCase(context.TODO(), &dataset.AddTestCaseRequest{
			Id:     tcs.DatasetID,
			Number: tcs.Number,
			Input:  tcs.Input,
			Output: tcs.Output,
		})
		if err != nil {
			return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to add test case #%d to dataset %s", tcs.Number, tcs.DatasetID)
		}
		log.Info("Test case successfully added")
	} else {
		needsUpdate, err := tcs.NeedsUpdate()
		if err != nil {
			return err
		}
		if needsUpdate {
			log.Info("Test case needs update, going to be updated")
			_, err = client.UpdateTestCase(context.TODO(), &dataset.UpdateTestCaseRequest{
				Id:     tcs.DatasetID,
				Number: tcs.Number,
				Input:  tcs.Input,
				Output: tcs.Output,
			})
			if err != nil {
				return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to update test case #%d to dataset %s", tcs.Number, tcs.DatasetID)
			}
			log.Info("Test case successfully updated")
		} else {
			log.Info("Test case is up to date, doing nothing")
		}
	}

	return nil
}
