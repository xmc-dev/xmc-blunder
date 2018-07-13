package db

import "github.com/google/uuid"

func (d *Datastore) SetPageID(model interface{}, id, pageID uuid.UUID) error {
	return e(d.db.Model(model).Where("id = ?", id).Update("page_id", pageID).Error, "couldn't set page id")
}
