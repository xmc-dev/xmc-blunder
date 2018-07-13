package login

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/auth-srv/handlers"
	"github.com/xmc-dev/xmc/auth-srv/session"
	"github.com/xmc-dev/xmc/auth-srv/util"
)

func CheckLoginHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.RequestURI()
		logrus.WithField("url", url).WithField("method", r.Method).Debug("Checking login handler")
		if strings.HasSuffix(url, "/login") || strings.HasSuffix(url, "/login/") {
			sessionCookie, err := r.Cookie("session_uuid")
			redirectURI := r.Form.Get("redirect_uri")
			if len(redirectURI) == 0 {
				redirectURI = "/"
			}

			// Cookie exists
			if err == nil {
				session, _ := session.ReadSession(sessionCookie.Value)
				if util.SessionOK(session) {
					logrus.WithField("redirectURI", redirectURI).Debug("Redirecting")
					http.Redirect(w, r, redirectURI, http.StatusTemporaryRedirect)
					return
				}
			}
		}

		// Cookie doesn't exist or session is invalid
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func GETHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	logrus.Debug("Url is ", r.URL.RequestURI())

	redirectURI := r.Form.Get("redirect_uri")
	if len(redirectURI) == 0 {
		redirectURI = "/"
	}

	t := template.Must(template.ParseFiles("static/login/index.html"))
	t.Execute(w, struct {
		RedirectURI string
	}{redirectURI})
}

func POSTHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	logrus.Debug("Url is ", r.URL.RequestURI())
	if !handlers.RequiredFields(w, r, "username", "password", "redirect_uri") {
		return
	}

	rsp, err := session.Login(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		if err.Code >= 500 && err.Code <= 599 {
			handlers.InternalServerError(w, err)
		} else if err.Code == http.StatusBadRequest {
			panic(err)
		} else {
			w.WriteHeader(int(err.Code))
			fmt.Fprintln(w, err.Detail)
		}
		return
	}

	cookie := http.Cookie{
		Name:     "session_uuid",
		Value:    rsp.Session.Uuid,
		Path:     "/",
		Expires:  time.Unix(rsp.Session.ExpiresAt, 0),
		HttpOnly: true,
	}

	handlers.SetCookie(w, r, &cookie)
	logrus.WithField("redirect_uri", r.Form.Get("redirect_uri")).Debug("Login success, redirecting...")
	http.Redirect(w, r, r.Form.Get("redirect_uri"), http.StatusSeeOther)
}
