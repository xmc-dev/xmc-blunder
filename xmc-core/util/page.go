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
	"github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
)

func CleanPagePath(p string) string {
	np := path.Clean("/" + strings.ToLower(p))
	np = strings.Replace(np, "/", ".", -1)

	return strings.TrimPrefix(np, ".")
}

func CreatePageVersion(d *db.Datastore, id uuid.UUID, timestamp time.Time, contents []byte, title string, attID uuid.UUID) error {
	ver := &mpage.Version{
		PageID:    id,
		Timestamp: timestamp,
		Title:     title,
	}

	var err error
	if contents != nil {
		attID, err = MakeAttachment(d, &attachment.CreateRequest{
			Attachment: &attachment.Attachment{
				ObjectId: "page/" + id.String(),
				Filename: timestamp.Format(time.RFC3339) + ".xmcml",
			},
			Contents: contents,
		})
		if err != nil {
			return err
		}
		err := d.SetAttachmentPublic(attID, true)
		if err != nil {
			return err
		}
	}
	ver.AttachmentID = attID
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
	err = CreatePageVersion(d, id, ts, req.Contents, req.Title, uuid.Nil)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func DeletePage(d *db.Datastore, id uuid.UUID, hard bool, log *logrus.Entry) error {
	vs, _, err := d.GetPageVersions(&page.GetVersionsRequest{Id: id.String()})
	if err != nil {
		return err
	}

	// delete all the versions' attachments
	if hard {
		for _, v := range vs {
			err := DeleteAttachment(d, v.AttachmentID)
			if err != nil {
				log.WithField("err", err).Warn("Error while deleting page version's attachment")
				return err
			}
		}
	}

	// The versions themselves will be deleted by db.DB.DeletePage in a transaction.
	// Btw if the page doesn't exist the code above that deleted the versions won't do anything.
	// The database is still queried tho, possible DoS vuln?
	// It might be if there are many versions in the table.
	err = d.DeletePage(id, hard)
	if err != nil {
		return err
	}

	return nil
}
