package util

import (
	"path"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/xmc-core/db"
	mpage "github.com/xmc-dev/xmc/xmc-core/db/models/page"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
)

func CleanPagePath(p string) string {
	np := path.Clean("/" + strings.ToLower(p))
	np = strings.Replace(np, "/", ".", -1)

	return strings.TrimPrefix(np, ".")
}

func CreatePageVersion(d *db.Datastore, id uuid.UUID, timestamp time.Time, content string, title string) error {
	ver := &mpage.Version{
		PageID:    id,
		Contents:  content,
		Timestamp: timestamp,
		Title:     title,
	}

	return d.CreatePageVersion(ver)
}

func CreatePage(d *db.Datastore, req *page.CreateRequest) (uuid.UUID, error) {
	req.Page.LatestTimestamp = ptypes.TimestampNow()
	req.Page.Id = ""
	req.Page.Path = CleanPagePath(req.Page.Path)
	id, ts, err := d.CreatePage(req.Page)
	if err != nil {
		return uuid.Nil, err
	}
	err = CreatePageVersion(d, id, ts, req.Contents, req.Title)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func DeletePage(d *db.Datastore, id uuid.UUID, hard bool, log *logrus.Entry) error {
	// The versions themselves will be deleted by db.DB.DeletePage in a transaction.
	err := d.DeletePage(id, hard)
	if err != nil {
		return err
	}

	return nil
}
