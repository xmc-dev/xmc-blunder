package handler

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/xmc-dev/xmc/common/perms"
	"github.com/xmc-dev/xmc/xmc-core/db"
	mpage "github.com/xmc-dev/xmc/xmc-core/db/models/page"
	"github.com/xmc-dev/xmc/xmc-core/proto/page"
	"github.com/xmc-dev/xmc/xmc-core/proto/searchmeta"
	"github.com/xmc-dev/xmc/xmc-core/util"
)

var pathRegex = regexp.MustCompile(`^([[:alnum:]]|_|/)+$`)

type PageService struct{}

func pageSName(method string) string {
	return fmt.Sprintf("%s.PageService.%s", "xmc.srv.core", method)
}

func (*PageService) isUpdatePointless(pg *mpage.Page, v *mpage.Version, req *page.UpdateRequest) bool {
	return (len(req.Title) == 0 || req.Title == v.Title) &&
		(p(req.Path) == pg.Path) &&
		(req.Contents == nil)
}

func p(p string) string {
	return util.CleanPagePath(p)
}

func isPathValid(p string) bool {
	return pathRegex.Match([]byte(p))
}

func (ps *PageService) Create(ctx context.Context, req *page.CreateRequest, rsp *page.CreateResponse) error {
	methodName := pageSName("Create")
	switch {
	case req.Page == nil:
		return errors.BadRequest(methodName, "missing page")
	case req.Contents == nil:
		return errors.BadRequest(methodName, "invalid contents")
	case len(req.Page.Path) == 0 || !isPathValid(req.Page.Path):
		return errors.BadRequest(methodName, "invalid path")
	case len(req.Title) == 0:
		return errors.BadRequest(methodName, "invalid title")
	}

	dd := db.DB.BeginGroup()
	id, err := util.CreatePage(dd, req)
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "path must be unique")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Id = id.String()
	return nil
}

func (*PageService) Read(ctx context.Context, req *page.ReadRequest, rsp *page.ReadResponse) error {
	methodName := pageSName("Read")
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	var ts *time.Time
	if req.Timestamp != nil {
		t, _ := ptypes.Timestamp(req.Timestamp)
		ts = &t
	}
	p, v, err := db.DB.ReadPage(id, ts)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "page or version not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Page = p.ToProto()
	rsp.Page.Version = v.ToProto()
	return nil
}

func (*PageService) Get(ctx context.Context, req *page.GetRequest, rsp *page.GetResponse) error {
	methodName := pageSName("Get")
	if len(req.Path) == 0 || !isPathValid(req.Path) {
		return errors.BadRequest(methodName, "invalid path")
	}

	p, v, err := db.DB.GetPage(p(req.Path))
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "page or version not found")
		}
		return errors.InternalServerError(methodName, e(err))
	}

	rsp.Page = p.ToProto()
	rsp.Page.Version = v.ToProto()
	return nil
}

func (*PageService) GetVersions(ctx context.Context, req *page.GetVersionsRequest, rsp *page.GetVersionsResponse) error {
	methodName := pageSName("GetVersions")
	_, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	vs, total, err := db.DB.GetPageVersions(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	vers := []*page.Version{}
	for _, v := range vs {
		vers = append(vers, v.ToProto())
	}

	rsp.Versions = vers
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(vers)),
		Total:   total,
	}
	return nil
}

func (*PageService) GetFirstChildren(ctx context.Context, req *page.GetFirstChildrenRequest, rsp *page.GetFirstChildrenResponse) error {
	methodName := pageSName("GetFirstChildren")
	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}
	if len(req.Id) == 0 {
		return errors.BadRequest(methodName, "invalid id")
	}
	ps, total, err := db.DB.GetFirstPageChildren(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	pages := []*page.Page{}
	for _, p := range ps {
		pages = append(pages, p.ToProto())
	}

	rsp.Pages = pages
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(pages)),
		Total:   total,
	}
	return nil
}

func (ps *PageService) Update(ctx context.Context, req *page.UpdateRequest, rsp *page.UpdateResponse) error {
	methodName := pageSName("Update")
	if !perms.HasScope(ctx, "manage/page") {
		return errors.Forbidden(methodName, "you are not allowed to update pages")
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	pg, v, err := dd.ReadPage(id, nil)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "page not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	// Don't forget to update this when you need
	if ps.isUpdatePointless(pg, v, req) {
		dd.Rollback()
		return nil
	}
	if len(req.Path) > 0 {
		if !isPathValid(req.Path) {
			dd.Rollback()
			return errors.BadRequest(methodName, "invalid path")
		}
		req.Path = p(req.Path)
	}
	var t *time.Time
	x := time.Now()
	t = &x
	title := req.Title
	if len(title) == 0 {
		title = v.Title
	}
	err = util.CreatePageVersion(dd, id, *t, req.Contents, title, v.AttachmentID)
	if err != nil {
		dd.Rollback()
		return errors.InternalServerError(methodName, "couldn't create page version")
	}

	err = dd.UpdatePage(req, t)
	if err != nil {
		dd.Rollback()
		if err == db.ErrUniqueViolation {
			return errors.Conflict(methodName, "path must be unique")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*PageService) Delete(ctx context.Context, req *page.DeleteRequest, rsp *page.DeleteResponse) error {
	methodName := pageSName("Delete")
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return errors.BadRequest(methodName, "invalid id")
	}

	dd := db.DB.BeginGroup()
	err = util.DeletePage(dd, id, req.Hard, log)
	if err != nil {
		dd.Rollback()
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "page not found")
		} else if e, ok := err.(db.ErrHasDependants); ok {
			return errors.BadRequest(methodName, "one or more "+string(e)+" depend on this page")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	if err := dd.Commit(); err != nil {
		return errors.InternalServerError(methodName, e(err))
	}

	return nil
}

func (*PageService) Search(ctx context.Context, req *page.SearchRequest, rsp *page.SearchResponse) error {
	methodName := pageSName("Search")

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 250 {
		req.Limit = 250
	}

	if len(req.Path) > 0 {
		req.Path = util.CleanPagePath(req.Path)
	}
	ps, total, err := db.DB.SearchPage(req)
	if err != nil {
		return errors.InternalServerError(methodName, e(err))
	}
	pages := []*page.Page{}
	for _, p := range ps {
		pages = append(pages, p.ToProto())
	}

	rsp.Pages = pages
	rsp.Meta = &searchmeta.Meta{
		PerPage: req.Limit,
		Count:   uint32(len(pages)),
		Total:   total,
	}
	return nil
}