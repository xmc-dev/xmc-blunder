package util

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"text/template"
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

func DirtyPagePath(p string) string {
	np := strings.Replace(p, ".", "/", -1)

	return "/" + np
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

type XMCTemplateError interface {
	error
	XMCTemplateError()
}

type ErrRecursiveTemplate struct {
	Path  string
	Trace []string
}

func (err ErrRecursiveTemplate) Error() string {
	var tr []string
	for _, t := range err.Trace {
		tr = append(tr, DirtyPagePath(t))
	}
	return fmt.Sprintf("recursive loop: %v", tr)
}

func (ErrRecursiveTemplate) XMCTemplateError() {}

type ErrBadMacroParams string

func (err ErrBadMacroParams) Error() string {
	return fmt.Sprintf("bad param: %s", string(err))
}

func (ErrBadMacroParams) XMCTemplateError() {}

type tmplDot struct {
	d        *db.Datastore
	included []string
}

func tmplInclude(path string, d tmplDot) (string, error) {
	found := false
	for _, t := range d.included {
		if t == CleanPagePath(path) {
			found = true
			break
		}
	}
	if found {
		return "", ErrRecursiveTemplate{
			Path:  CleanPagePath(path),
			Trace: d.included,
		}
	}
	b := &strings.Builder{}
	err := RenderPage(d.d, path, b, d.included)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func tmplMacro(name string, params ...string) (string, error) {
	const delim = "="
	newParams := map[string]interface{}{}
	for _, p := range params {
		split := strings.Split(p, delim)
		// 2 for a key and a value
		if len(split) < 2 {
			return "", ErrBadMacroParams(p)
		}
		val := strings.Join(split[1:], delim)
		var iface interface{}
		if err := json.Unmarshal([]byte(val), &iface); err != nil {
			if err := json.Unmarshal([]byte(`"`+val+`"`), &iface); err != nil {
				return "", err
			}
		}
		newParams[split[0]] = iface
	}
	buf, err := json.Marshal(newParams)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("!!%s%s!!", name, string(buf)), nil
}

func RenderPage(d *db.Datastore, path string, w io.Writer, previous []string) error {
	clean := CleanPagePath(path)
	_, v, err := d.GetPage(clean)
	if err != nil {
		return err
	}
	t := template.New(clean).Funcs(template.FuncMap{
		"include": tmplInclude,
		"macro":   tmplMacro,
	})
	t, err = t.Parse(v.Contents)
	if err != nil {
		return err
	}
	return t.Execute(w, tmplDot{
		d:        d,
		included: append(previous, clean),
	})
}

func RenderAsString(d *db.Datastore, path string) (string, error) {
	b := &strings.Builder{}
	err := RenderPage(d, path, b, []string{})
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
