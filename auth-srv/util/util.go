package util

import (
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/account-srv/proto/session"
)

func WarnErr(err error, params ...interface{}) {
	logrus.WithField("err", err).Warn(params...)
}

func SessionOK(s *session.Session) bool {
	return s != nil && time.Now().Unix() < s.ExpiresAt
}

func HasScopes(scopeList string, scopes string) bool {
	sl := make(map[string]struct{})

	for _, scope := range strings.Split(scopeList, " ") {
		sl[scope] = struct{}{}
	}

	for _, scope := range strings.Split(scopes, " ") {
		_, ok := sl[scope]
		if !ok {
			return false
		}
	}

	return true
}

func Redirect(w http.ResponseWriter, r *http.Request, uri string, code int) {
	prefix := path.Clean(r.Header.Get("X-Forwarded-Prefix"))
	path := ""
	if prefix == "." {
		path = uri
	} else {
		path = prefix + "/" + uri
	}
	http.Redirect(w, r, path, code)
}
