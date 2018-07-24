package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/xmc-dev/xmc/xmc-core/db/models/page"
	ppage "github.com/xmc-dev/xmc/xmc-core/proto/page"
)

func (d *Datastore) updateLastPageEventInternal() error {
	return d.db.Exec("UPDATE internal SET last_page_event = ?", time.Now()).Error
}

func (d *Datastore) CreatePage(pg *ppage.Page) (uuid.UUID, time.Time, error) {
	dd := d.begin()
	p := page.FromProto(pg)
	p.CreatedAt = p.LatestTimestamp
	err := dd.db.Create(p).Error
	if err != nil {
		dd.Rollback()
		return uuid.Nil, time.Time{}, e(err, "couldn't create page")
	}
	err = dd.updateLastPageEventInternal()
	if err != nil {
		dd.Rollback()
		return uuid.Nil, time.Time{}, e(err, "couldn't create page")
	}
	err = dd.Commit()
	if err != nil {
		return uuid.Nil, time.Time{}, e(err, "couldn't create page")
	}
	return p.ID, p.CreatedAt, nil
}

func (d *Datastore) CreatePageVersion(ver *page.Version) error {
	return e(d.db.Create(ver).Error, "couldn't create page version")
}

func (d *Datastore) ReadPageVersion(pageID uuid.UUID, timestamp *time.Time) (*page.Version, error) {
	query := d.db
	if timestamp != nil {
		query = query.Where("timestamp = ?", *timestamp)
	}

	v := &page.Version{}
	err := query.Order("timestamp DESC").First(v, "page_id = ?", pageID).Error
	return v, e(err, "couldn't read page version")
}

func (d *Datastore) ReadPage(id uuid.UUID, timestamp *time.Time) (*page.Page, *page.Version, error) {
	dd := d.begin()
	p := &page.Page{}

	err := dd.db.First(p, "id = ?", id).Error
	if err != nil {
		dd.Rollback()
		return nil, nil, e(err, "couldn't read page")
	}

	v, err := dd.ReadPageVersion(id, timestamp)
	if err != nil {
		dd.Rollback()
		return nil, nil, err
	}
	return p, v, e(dd.Commit(), "couldn't read page")
}

func (d *Datastore) GetPage(path string) (*page.Page, *page.Version, error) {
	dd := d.begin()
	p := &page.Page{}

	err := dd.db.First(p, "path = ?", path).Error
	if err != nil {
		dd.Rollback()
		return nil, nil, e(err, "couldn't get page")
	}

	v, err := dd.ReadPageVersion(p.ID, nil)
	if err != nil {
		dd.Rollback()
		return nil, nil, err
	}
	return p, v, e(dd.Commit(), "couldn't get page")
}

func (d *Datastore) UpdatePage(pg *ppage.UpdateRequest, latestTimestamp *time.Time) error {
	dd := d.begin()
	id, _ := uuid.Parse(pg.Id)

	p, _, err := dd.ReadPage(id, nil)
	if err != nil {
		dd.Rollback()
		return err
	}

	if len(pg.Path) > 0 {
		p.Path = pg.Path
	}
	if latestTimestamp != nil {
		p.LatestTimestamp = *latestTimestamp
	}

	if err := dd.db.Save(p).Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't update page")
	}

	return e(dd.Commit(), "couldn't update page")
}

func (d *Datastore) DeletePage(id uuid.UUID, hard bool) error {
	dd := d.begin()
	q := dd.db
	if hard {
		q = q.Unscoped()
	}
	result := q.Where("id = ?", id).Delete(&page.Page{})
	if err := result.Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't delete page")
	}
	if result.RowsAffected == 0 {
		dd.Rollback()
		return ErrNotFound
	}
	if hard {
		if err := q.Where("page_id = ?", id).Delete(&page.Version{}).Error; err != nil {
			dd.Rollback()
			return e(err, "couldn't delete page versions")
		}
	}
	err := dd.updateLastPageEventInternal()
	if err != nil {
		dd.Rollback()
		return e(err, "couldn't delete page")
	}
	err = dd.Commit()
	if err != nil {
		return e(err, "couldn't delete page")
	}
	return nil
}

func (d *Datastore) UndeletePage(id uuid.UUID) error {
	dd := d.begin()
	q := dd.db.Unscoped()
	result := q.Model(&page.Page{}).Where("id = ?", id).Update("deleted_at", nil)
	if err := result.Error; err != nil {
		dd.Rollback()
		return e(err, "couldn't undelete page")
	}
	if result.RowsAffected == 0 {
		dd.Rollback()
		return ErrNotFound
	}
	err := dd.updateLastPageEventInternal()
	if err != nil {
		dd.Rollback()
		return e(err, "couldn't undelete page")
	}
	err = dd.Commit()
	if err != nil {
		return e(err, "couldn't undelete page")
	}
	return nil
}

func (d *Datastore) SearchPage(req *ppage.SearchRequest) ([]*page.Page, uint32, error) {
	dd := d.begin()
	ps := []*page.Page{}
	query := dd.db.Order("created_at ASC").Unscoped()
	if len(req.Path) > 0 {
		query = query.Where("text(path) ~* ?", req.Path)
	}
	if len(req.Title) > 0 {
		query = query.Where(
			"(SELECT title FROM page_versions WHERE page_id = pages.id ORDER BY timestamp DESC LIMIT 1) ~* ?",
			req.Title,
		)
	}
	if len(req.ObjectId) > 0 {
		query = query.Where("object_id = ?", req.ObjectId)
	}
	var cnt uint32
	if err := query.Model(&ps).Count(&cnt).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search pages")
	}

	query = query.Limit(req.Limit).Offset(req.Offset)
	if err := query.Find(&ps).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't search pages")
	}

	err := dd.Commit()

	return ps, cnt, e(err, "couldn't search pages")
}

func (d *Datastore) GetPageVersions(req *ppage.GetVersionsRequest) ([]*page.Version, uint32, error) {
	dd := d.begin()
	vs := []*page.Version{}
	var cnt uint32
	if err := dd.db.Model(&vs).Where("page_id = ?", req.Id).Count(&cnt).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't get page versions")
	}
	query := dd.db.Offset(req.Offset)
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	query = query.Where("page_id = ?", req.Id)
	if err := query.Find(&vs).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't get page versions")
	}

	err := dd.Commit()

	return vs, cnt, e(err, "couldn't get page versions")
}

func (d *Datastore) GetFirstPageChildren(req *ppage.GetFirstChildrenRequest) ([]*page.Page, uint32, error) {
	dd := d.begin()
	ps := []*page.Page{}
	if err := dd.db.Exec("SELECT xmc_page_children()").Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't get page's first children")
	}
	query := dd.db.Where("parent_id = ?", req.Id)
	var cnt uint32
	if err := query.Model(&ps).Count(&cnt).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't get page's first children")
	}

	query = query.Limit(req.Limit).Offset(req.Offset)
	if err := query.Find(&ps).Error; err != nil {
		dd.Rollback()
		return nil, 0, e(err, "couldn't get page's first children")
	}

	err := dd.Commit()

	return ps, cnt, e(err, "couldn't get page's first children")
}
