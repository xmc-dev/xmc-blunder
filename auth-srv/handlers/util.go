package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/auth-srv/util"
)

// HTTPError returns an HTTP error
func HTTPError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	fmt.Fprintln(w, msg)
	logrus.WithFields(logrus.Fields{
		"code": code,
		"msg":  msg,
	}).Debug("HTTPError")
}

// RequiredFields checks if the given fields are present and returns an http error message if not.
// Returns false if one of the required field is missing
func RequiredFields(w http.ResponseWriter, r *http.Request, fields ...string) bool {
	for _, f := range fields {
		if len(r.Form[f]) == 0 {
			HTTPError(w, http.StatusBadRequest, f+" cannot be empty")
			return false
		}
	}

	return true
}

func InternalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	util.WarnErr(err, "Internal server error")
}

func SetCookie(w http.ResponseWriter, r *http.Request, cookie *http.Cookie) {
	if strings.HasPrefix(r.Host, "localhost") {
		cookie.Secure = false
	} else {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func RequiredCookies(w http.ResponseWriter, r *http.Request, cookies ...string) bool {
	for _, c := range cookies {
		_, err := r.Cookie(c)
		if err == http.ErrNoCookie {
			HTTPError(w, http.StatusBadRequest, "cookie '"+c+"' cannot be empty")
			return false
		}
	}

	return true
}

func DeleteCookie(name string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	return cookie
}
