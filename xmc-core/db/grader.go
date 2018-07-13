package db

import (
	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/xmc-core/db/models/problem"
	pgrader "github.com/xmc-dev/xmc/xmc-core/proto/grader"
)

func (d *Datastore) CreateGrader(gd *pgrader.Grader) (uuid.UUID, error) {
	g := problem.GraderFromProto(gd)

	err := d.db.Create(g).Error
	return g.ID, e(err, "couldn't create grader")
}

func (d *Datastore) ReadGrader(id uuid.UUID) (*problem.Grader, error) {
	g := &problem.Grader{}

	err := d.db.Where("id = ?", id).First(g).Error

	return g, e(err, "couldn't read grader")
}

func (d *Datastore) GetGrader(name string) (*problem.Grader, error) {
	g := &problem.Grader{}

	err := d.db.Where("name = ?", name).First(g).Error

	return g, e(err, "couldn't get grader by name")
}

func (d *Datastore) UpdateGrader(gd *pgrader.UpdateRequest) error {
	dd := d.begin()
	id, _ := uuid.Parse(gd.Id)
	g, err := dd.ReadGrader(id)
	if err != nil {
		dd.Rollback()
		return err
	}

	if len(gd.Language) > 0 {
		g.Language = gd.Language
	}
	if len(gd.Name) > 0 {
		g.Name = gd.Name
	}

	if err := dd.db.Save(g).Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't update grader")
	}

	return e(dd.Commit(), "couldn't upgrade grader")
}

func (d *Datastore) DeleteGrader(id uuid.UUID) error {
	result := d.db.Where("id = ?", id).Delete(&problem.Grader{})
	if result.Error != nil {
		return e(result.Error, "couldn't delete grader")
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *Datastore) GraderSetAttachmentID(id uuid.UUID, attID uuid.UUID) error {
	return e(d.db.Exec("UPDATE graders SET attachment_id = ? WHERE id = ?", attID, id).Error, "couldn't set grader's attachment id")
}

func (d *Datastore) SearchGrader(req *pgrader.SearchRequest) ([]*problem.Grader, uint32, error) {
	dd := d.begin()
	gs := []*problem.Grader{}

	query := dd.db
	if len(req.Name) > 0 {
		query = query.Where("name ~* ?", req.Name)
	}
	if len(req.Language) > 0 {
		query = query.Where("language ~* ?", req.Language)
	}

	var cnt uint32
	if err := query.Model(&gs).Count(&cnt).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search graders")
	}

	query = query.Limit(req.Limit).Offset(req.Offset)
	if err := query.Find(&gs).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search graders")
	}

	err := dd.Commit()
	return gs, cnt, e(err, "couldn't search graders")
}
