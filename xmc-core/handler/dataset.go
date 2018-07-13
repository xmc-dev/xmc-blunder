package handler

import (
	"bytes"
	"context"
	"fmt"

	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/xmc-core/db"
	"github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	"github.com/xmc-dev/xmc/xmc-core/proto/dataset"
	"github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"
	"github.com/xmc-dev/xmc/xmc-core/s3"
	"github.com/xmc-dev/xmc/xmc-core/util"
)

type DatasetService struct{}

func datasetSName(method string) string {
	return fmt.Sprintf("%s.DatasetService.%s", "xmc.srv.core", method)
}

func (*DatasetService) Create(ctx context.Context, req *dataset.CreateRequest, rsp *dataset.CreateResponse) error {
	methodName := datasetSName("Create")
	switch {
	case req.Dataset == nil:
		return errors.BadRequest(methodName, "missing dataset")
	case len(req.Dataset.Name) == 0:
		return errors.BadRequest(methodName, "invalid name")
	case len(req.Dataset.GraderId) == 0:
		return errors.BadRequest(methodName, "invalid grader_id")
	case req.Dataset.MemoryLimit == 0:
		return errors.BadRequest(methodName, "invalid memory_limit")
	case req.Dataset.TimeLimit == nil:
		return errors.BadRequest(methodName, "invalid time_limit")
	}

	_, err := ptypes.Duration(req.Dataset.TimeLimit)
	if err != nil {
		return errors.BadRequest(methodName, "invalid time_limit")
	}

	req.Dataset.Name = strings.ToLower(req.Dataset.Name)

	id, err := db.DB.CreateDataset(req.Dataset)
	if err != nil {
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		} else if _, ok := err.(db.ErrHasDependants); ok {
			return errors.BadRequest(methodName, "grader doesn't exist")
		}
		return errors.InternalServerError(methodName, e(err))
	}
	rsp.Id = id.String()
	return nil
}

func (*DatasetService) Read(ctx context.Context, req *dataset.ReadRequest, rsp *dataset.ReadResponse) error {
	methodName := datasetSName("Read")
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid id")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	d, err := db.DB.ReadDataset(id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "dataset not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Dataset = d.ToProto()
	return nil
}

func (*DatasetService) Get(ctx context.Context, req *dataset.GetRequest, rsp *dataset.GetResponse) error {
	methodName := datasetSName("Get")
	if len(req.Name) == 0 {
		return errors.BadRequest(methodName, "invalid name")
	}
	req.Name = strings.ToLower(req.Name)

	d, err := db.DB.GetDataset(req.Name)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "dataset not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Dataset = d.ToProto()
	return nil
}

func (*DatasetService) Update(ctx context.Context, req *dataset.UpdateRequest, rsp *dataset.UpdateResponse) error {
	methodName := datasetSName("Update")
	_, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	if len(req.GraderId) > 0 {
		graderID, err := uuid.Parse(req.GraderId)
		if err != nil {
			dd.Rollback()
			return errors.BadRequest(methodName, "invalid grader_id")
		}
		_, err = dd.ReadGrader(graderID)
		if err != nil {
			dd.Rollback()
			if err == db.ErrNotFound {
				return errors.BadRequest(methodName, "grader does not exist")
			}
			return errors.InternalServerError(methodName, e(err))
		}
	}

	err = dd.UpdateDataset(req)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "dataset not found")
		} else if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "name must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	return nil
}

func (*DatasetService) Delete(ctx context.Context, req *dataset.DeleteRequest, rsp *dataset.DeleteResponse) error {
	methodName := datasetSName("Delete")
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid id")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	tcs, _, err := dd.ReadTestCases(id)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.DeleteDataset(id)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "dataset not found")
		} else if e, ok := err.(db.ErrHasDependants); ok {
			return errors.BadRequest(methodName, "one or more "+string(e)+" depend on this dataset")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	for _, tc := range tcs {
		err = util.DeleteAttachment(db.DB, tc.InputAttachmentID)
		if err != nil {
			log.Warnf("Error while deleting test case's input file attachment id %v: %v", tc.InputAttachmentID, err)
		}
		err = util.DeleteAttachment(db.DB, tc.OutputAttachmentID)
		if err != nil {
			log.Warnf("Error while deleting test case's output file attachment id %v: %v", tc.OutputAttachmentID, err)
		}
	}
	return nil
}

func (*DatasetService) Search(ctx context.Context, req *dataset.SearchRequest, rsp *dataset.SearchResponse) error {
	methodName := datasetSName("Search")

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	if len(req.GraderId) > 0 {
		_, err := uuid.Parse(req.GraderId)
		if err != nil {
			return errors.BadRequest(methodName, "invalid grader_uuid")
		}
	}
	dts, total, err := db.DB.SearchDataset(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	ds := []*dataset.Dataset{}
	for _, dt := range dts {
		ds = append(ds, dt.ToProto())
	}

	rsp.Datasets = ds
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(ds)),
		Total:   total,
	}
	return nil
}

func (*DatasetService) AddTestCase(ctx context.Context, req *dataset.AddTestCaseRequest, rsp *dataset.AddTestCaseResponse) error {
	methodName := datasetSName("AddTestCase")
	switch {
	case req.Number == 0:
		return errors.BadRequest(methodName, "invalid number")
	case req.Input == nil:
		return errors.BadRequest(methodName, "invalid input")
	case req.Output == nil:
		return errors.BadRequest(methodName, "invalid output")
	}

	// id is the dataset id
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	_, cnt, err := dd.ReadTestCases(id)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}
	if uint32(req.Number) != cnt+1 {
		dd.Rollback()
		return errors.BadRequest(methodName, "test cases must have consecutive numbers")
	}

	testCaseID, err := dd.CreateTestCase(req)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.BadRequest(methodName, "dataset doesn't exist")
		} else if err == db.ErrUniqueViolation {
			return errors.BadRequest(methodName, "a testcase of the same dataset with the same number already exists")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	inputID, err := util.MakeAttachment(dd, &attachment.CreateRequest{
		Attachment: &attachment.Attachment{
			ObjectId: "testcase/" + testCaseID.String(),
			Filename: fmt.Sprintf("test%d.in", req.Number),
		},
		Contents: req.Input,
	})
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	outputID, err := util.MakeAttachment(dd, &attachment.CreateRequest{
		Attachment: &attachment.Attachment{
			ObjectId: "testcase/" + testCaseID.String(),
			Filename: fmt.Sprintf("test%d.ok", req.Number),
		},
		Contents: req.Output,
	})
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	err = dd.TestCaseSetAttachmentIDs(testCaseID, inputID, outputID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	return nil
}

func (*DatasetService) GetTestCases(ctx context.Context, req *dataset.GetTestCasesRequest, rsp *dataset.GetTestCasesResponse) error {
	methodName := datasetSName("GetTestCases")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	ts, _, err := db.DB.ReadTestCases(id)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	tcs := []*dataset.TestCase{}
	for _, t := range ts {
		tcs = append(tcs, t.ToProto())
	}

	rsp.TestCases = tcs

	return nil
}

func (*DatasetService) GetTestCase(ctx context.Context, req *dataset.GetTestCaseRequest, rsp *dataset.GetTestCaseResponse) error {
	methodName := datasetSName("GetTestCase")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	t, err := db.DB.ReadTestCase(id, req.Number)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "test case not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.TestCase = t.ToProto()

	return nil
}

func (*DatasetService) UpdateTestCase(ctx context.Context, req *dataset.UpdateTestCaseRequest, rsp *dataset.UpdateTestCaseResponse) error {
	methodName := datasetSName("UpdateTestCase")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	t, err := dd.ReadTestCase(id, req.Number)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "test case doesn't exist")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if req.Input != nil {
		att, err := dd.ReadAttachment(t.InputAttachmentID)
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}

		reader := bytes.NewReader(req.Input)
		_, err = s3.UploadAttachment(att, reader, int64(len(req.Input)))
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
	}

	if req.Output != nil {
		att, err := dd.ReadAttachment(t.OutputAttachmentID)
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}

		reader := bytes.NewReader(req.Output)
		_, err = s3.UploadAttachment(att, reader, int64(len(req.Output)))
		if err != nil {
			dd.Rollback()
			return errors.InternalServerError(methodName, e(err))
		}
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*DatasetService) RemoveTestCase(ctx context.Context, req *dataset.RemoveTestCaseRequest, rsp *dataset.RemoveTestCaseResponse) error {
	methodName := datasetSName("RemoveTestCase")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	if req.Number == 0 {
		return errors.BadRequest(methodName, "invalid number")
	}

	err = db.DB.RemoveTestCase(id, req.Number)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "test case not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}
