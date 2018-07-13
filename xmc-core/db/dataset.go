package db

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/xmc-core/db/models/problem"
	pdataset "github.com/xmc-dev/xmc/xmc-core/proto/dataset"
)

func (d *Datastore) CreateDataset(ds *pdataset.Dataset) (uuid.UUID, error) {
	dt := problem.DatasetFromProto(ds)

	err := d.db.Create(dt).Error
	return dt.ID, e(err, "couldn't create dataset")
}

func (d *Datastore) ReadDataset(uuid uuid.UUID) (*problem.Dataset, error) {
	dt := &problem.Dataset{}

	err := d.db.First(dt, "id = ?", uuid).Error
	return dt, e(err, "couldn't read dataset")
}

func (d *Datastore) GetDataset(name string) (*problem.Dataset, error) {
	dt := &problem.Dataset{}

	err := d.db.First(dt, "name = ?", name).Error
	return dt, e(err, "couldn't get dataset by name")
}

func (d *Datastore) UpdateDataset(ds *pdataset.UpdateRequest) error {
	dd := d.begin()
	id, _ := uuid.Parse(ds.Id)
	dt, err := dd.ReadDataset(id)
	if err != nil {
		dd.Rollback()
		return err
	}

	if len(ds.Description) > 0 {
		dt.Description = ds.Description
	}

	if len(ds.GraderId) > 0 {
		graderID, _ := uuid.Parse(ds.GraderId)
		dt.GraderID = graderID
	}

	if len(ds.Name) > 0 {
		dt.Name = ds.Name
	}

	if ds.MemoryLimit > 0 {
		dt.MemoryLimit = ds.MemoryLimit
	}

	if ds.TimeLimit != nil {
		dt.TimeLimit, _ = ptypes.Duration(ds.TimeLimit)
	}

	if err := dd.db.Save(dt).Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't update dataset")
	}

	return e(dd.Commit(), "couldn't update dataset")
}

func (d *Datastore) DeleteDataset(id uuid.UUID) error {
	result := d.db.Where("id = ?", id).Delete(&problem.Dataset{})
	if result.Error != nil {
		return e(result.Error, "couldn't delete dataset")
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *Datastore) DatasetExists(uuid uuid.UUID) (bool, error) {
	row := d.db.Raw("SELECT EXISTS(SELECT 1 FROM datasets)").Row()
	result := false

	err := row.Scan(&result)

	return result, e(err, "couldn't check for dataset existence")
}

func (d *Datastore) SearchDataset(req *pdataset.SearchRequest) ([]*problem.Dataset, uint32, error) {
	dd := d.begin()
	ds := []*problem.Dataset{}
	query := dd.db
	if len(req.Description) > 0 {
		query = query.Where("description ~* ?", req.Description)
	}

	if len(req.GraderId) > 0 {
		graderID, _ := uuid.Parse(req.GraderId)
		query = query.Where("grader_id = ?", graderID)
	}

	if len(req.Name) > 0 {
		query = query.Where("name ~* ?", req.Name)
	}
	var cnt uint32
	err := query.Model(&ds).Count(&cnt).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search datasets")
	}
	query = query.Limit(req.Limit).Offset(req.Offset)
	err = query.Find(&ds).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search datasets")
	}

	err = dd.Commit()
	return ds, cnt, e(err, "couldn't search datasets")
}

func (d *Datastore) CreateTestCase(req *pdataset.AddTestCaseRequest) (uuid.UUID, error) {
	datasetID, _ := uuid.Parse(req.Id)
	t := &problem.TestCase{
		Number:    req.Number,
		DatasetID: datasetID,
	}

	err := d.db.Create(t).Error

	return t.ID, e(err, "couldn't create test case")
}

func (d *Datastore) ReadTestCases(datasetID uuid.UUID) ([]*problem.TestCase, uint32, error) {
	dd := d.begin()
	ts := []*problem.TestCase{}
	query := dd.db.Where("dataset_id = ?", datasetID).Order("number ASC")

	var cnt uint32
	err := query.Model(&problem.TestCase{}).Count(&cnt).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't read test cases")
	}
	err = query.Find(&ts).Error
	if err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't read test cases")
	}

	err = dd.Commit()
	return ts, cnt, e(err, "couldn't read test cases")
}

func (d *Datastore) ReadTestCase(datasetID uuid.UUID, number int32) (*problem.TestCase, error) {
	t := &problem.TestCase{}

	err := d.db.First(t, "dataset_id = ? AND number = ?", datasetID, number).Error
	return t, e(err, "couldn't read test case")
}

func (d *Datastore) DeleteTestCase(testCaseID uuid.UUID) error {
	return e(d.db.Where("id = ?", testCaseID).Error, "couldn't delete test case")
}

func (d *Datastore) TestCaseSetAttachmentIDs(testCaseID, inputID, outputID uuid.UUID) error {
	return e(d.db.Exec("UPDATE test_cases SET	input_attachment_id = ?, output_attachment_id = ? WHERE id = ?", inputID, outputID, testCaseID).Error, "couldn't set test case's attachment ids")
}

func (d *Datastore) RemoveTestCase(datasetID uuid.UUID, number int32) error {
	result := d.db.Where("dataset_id = ? AND number = ?", datasetID, number).Delete(&problem.TestCase{})

	if result.Error != nil {
		return e(result.Error, "couldn't remove test case")
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
