package xmc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/xmc-dev/xmc/xmc-core/common"
	"github.com/xmc-dev/xmc/xmc-core/importer"
)

// GraderImporter imports graders stored in the XMC format.
//
// In the XMC format, the grader is an ordinary text source code file
// with the filename of the form gradername.ext, where ext is the typical
// file extension for the programming language the grader is written in.
type GraderImporter struct {
}

func NewGraderImporter() *GraderImporter {
	return &GraderImporter{}
}

func (gi *GraderImporter) ReadGrader(fp string) (*importer.GraderSpec, error) {
	fi, err := os.Stat(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-grader-importer: couldn't stat file %s", fp)
	}
	if fi.IsDir() {
		return nil, errors.New("xmc-grader-importer: invalid grader, path " + fp + " is a directory")
	}

	gs := &importer.GraderSpec{}
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "xmc-grader-importer: failed to read grader")
	}
	gs.Contents = b
	gs.Language, err = common.FileExtToLanguage(filepath.Ext(fp)[1:])
	if err != nil {
		return nil, errors.New("xmc-grader-importer: invalid language extension")
	}
	gs.Name = strings.ToLower(strings.TrimSuffix(filepath.Base(fp), filepath.Ext(fp)))

	return gs, nil
}
