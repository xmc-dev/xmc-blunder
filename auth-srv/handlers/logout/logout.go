package logout

import (
	"net/http"

	"github.com/xmc-dev/xmc/auth-srv/handlers"
	"github.com/xmc-dev/xmc/auth-srv/session"
	"github.com/xmc-dev/xmc/auth-srv/util"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	redirectURI := r.Form.Get("redirect_uri")
	if len(redirectURI) == 0 {
		redirectURI = "/"
	}

	sessionCookie, err := r.Cookie("session_uuid")
	if err == nil {
		_, err := session.Logout(sessionCookie.Value)
		if err != nil {
			if err.Code >= 500 && err.Code <= 599 {
				util.WarnErr(err)
			} else if err.Code == http.StatusBadRequest {
				panic(err)
			}
		}
	}

	handlers.SetCookie(w, r, handlers.DeleteCookie("session_uuid"))
	http.Redirect(w, r, redirectURI, http.StatusTemporaryRedirect)
}
