package importer

import (
	"context"

	merrors "github.com/micro/go-micro/errors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/xmc-core/common"
	"github.com/xmc-dev/xmc/xmc-core/proto/grader"
)

// GraderSpec is a generic representation of the soon-to-be-imported Grader
type GraderSpec struct {
	Name     string
	Language common.Language
	Contents []byte

	graderID string
}

// GraderImporter imports graders from the filesystem according to a defined format
type GraderImporter interface {
	ReadGrader(fp string) (*GraderSpec, error)
}

func graderClient() grader.GraderServiceClient {
	return grader.NewGraderServiceClient("xmc.srv.core", BaseClient)
}

func (gs *GraderSpec) getGraderID() error {
	if len(gs.graderID) > 0 {
		return nil
	}

	log := Log.WithFields(logrus.Fields{
		"object":      "grader",
		"grader_name": gs.Name,
	})
	var err error

	gs.graderID, err = getGraderID(gs.Name, log)

	return err
}

func (gs *GraderSpec) IsNew() (bool, error) {
	err := gs.getGraderID()
	switch e := errors.Cause(err).(type) {
	case nil:
		return false, nil
	case *ErrNotFound:
		return true, nil
	default:
		return false, e
	}
}

func (gs *GraderSpec) NeedsUpdate() (bool, error) {
	return true, nil
}

func (gs *GraderSpec) Import() error {
	client := graderClient()
	log := Log.WithFields(logrus.Fields{
		"object":      "grader",
		"grader_name": gs.Name,
	})

	isNew, err := gs.IsNew()
	if err != nil {
		return err
	}

	if isNew {
		log.Info("Grader is new, going to be created")
		_, err := client.Create(context.TODO(), &grader.CreateRequest{
			Grader: &grader.Grader{
				Language: string(gs.Language),
				Name:     gs.Name,
			},
			Code: gs.Contents,
		})
		if err != nil {
			return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to create grader %s", gs.Name)
		}
		log.Info("Grader successfully created")
	} else {
		needsUpdate, err := gs.NeedsUpdate()
		if err != nil {
			return err
		}
		if needsUpdate {
			log.Info("Grader needs update, going to be updated")
			err = gs.getGraderID()
			if err != nil {
				return err
			}
			_, err = client.Update(context.TODO(), &grader.UpdateRequest{
				Id:       gs.graderID,
				Code:     gs.Contents,
				Language: string(gs.Language),
				Name:     gs.Name,
			})
			if err != nil {
				return errors.Wrapf(merrors.Parse(err.Error()), "importer: failed to update grader %s", gs.Name)
			}
			log.Info("Grader successfully updated")
		} else {
			log.Info("Grader is up to date, doing nothing")
		}
	}

	return nil
}
