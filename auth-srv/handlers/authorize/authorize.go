package authorize

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strings"

	paccount "github.com/xmc-dev/xmc/account-srv/proto/account"
	psession "github.com/xmc-dev/xmc/account-srv/proto/session"
	"github.com/xmc-dev/xmc/auth-srv/account"
	"github.com/xmc-dev/xmc/auth-srv/handlers"
	csession "github.com/xmc-dev/xmc/auth-srv/session"
	"github.com/xmc-dev/xmc/auth-srv/storage"
	"github.com/xmc-dev/xmc/auth-srv/util"

	"github.com/RangelReale/osin"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/auth-srv/globals"
)

func getOwner(uuid string) (*paccount.Account, error) {
	rsp, err := account.ReadAccount(uuid)

	if err != nil {
		return nil, err
	}

	return rsp.Account, nil
}

func getConsentRequestData(ar *osin.AuthorizeRequest, subjectID string) *util.ConsentRequestData {
	crd := util.ConsentRequestData{}
	client, ok := ar.Client.(*storage.Client)
	if !ok {
		return nil
	}
	owner, err := getOwner(client.Original.OwnerUuid)
	if err != nil {
		return nil
	}
	rawSubject, err := globals.OsinServer.Storage.GetClient(subjectID)
	if err != nil {
		return nil
	}
	subject, ok := rawSubject.(*storage.Client)
	if !ok {
		return nil
	}

	crd.ClientName = client.Original.Name
	crd.OwnerClientID = owner.ClientId
	crd.SubjectClientID = subject.Original.ClientId
	if ar.Type == osin.CODE {
		crd.ResponseType = "code"
	} else {
		crd.ResponseType = "token"
	}
	crd.ClientID = client.Original.ClientId
	crd.RedirectURI = ar.RedirectUri
	crd.State = ar.State
	crd.Scope = ar.Scope
	crd.Prefix = path.Clean(ar.HttpRequest.Header.Get("X-Forwarded-Prefix"))
	if crd.Prefix == "." {
		crd.Prefix = ""
	}
	crd.CodeChallenge = ar.CodeChallenge
	crd.CodeChallengeMethod = ar.CodeChallengeMethod

	for _, s := range strings.Split(ar.Scope, " ") {
		crd.Scopes = append(crd.Scopes, util.DetailedScope{
			Title:    s,
			Synopsis: "Dummy text",
		})
	}

	return &crd
}

func Handler(w http.ResponseWriter, r *http.Request) {
	resp := globals.OsinServer.NewResponse()
	defer resp.Close()

	if ar := globals.OsinServer.HandleAuthorizeRequest(resp, r); ar != nil {
		sessionCookie, err := r.Cookie("session_uuid")
		var session *psession.Session
		var serr error
		if err == nil && sessionCookie != nil {
			session, serr = csession.ReadSession(sessionCookie.Value)
		}
		logrus.WithFields(logrus.Fields{
			"cookie":  sessionCookie,
			"session": session,
			"serr":    serr,
		}).Debug("Got authorization info")

		// if session is invalid then we need to authenticate the user first
		if err != nil || sessionCookie == nil ||
			!util.SessionOK(session) {
			reqURI := r.Header.Get("X-Forwarded-Prefix")
			if reqURI[len(reqURI)-1] == '/' {
				reqURI = reqURI[:len(reqURI)-1]
			}
			reqURI += r.URL.RequestURI()
			util.Redirect(w, r, "/login/?redirect_uri="+
				url.QueryEscape(reqURI), http.StatusSeeOther)
			return
		}

		creq := &storage.ConsentRequest{
			Scope:    ar.Scope,
			State:    ar.State,
			ClientID: ar.Client.GetId(),
		}

		logrus.WithField("authorize", r.Form.Get("authorize")).Debug("Authorizing")
		if r.Form.Get("authorize") == "1" {
			// User pressed accept button
			ok := globals.ConsentStorage.ValidateRequest(creq)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		} else {
			// Present consent page to user
			err := globals.ConsentStorage.SaveRequest(creq)
			if err != nil {
				handlers.InternalServerError(w, err)
			} else {
				t := template.Must(template.ParseFiles("static/consent/index.html"))
				crd := getConsentRequestData(ar, session.ClientId)
				logrus.WithField("crd", fmt.Sprintf("%+v", crd)).Debug("Got CRD")
				t.Execute(w, crd)
			}
			return
		}

		ar.Authorized = true
		ar.Expiration = 32000
		ar.UserData = &storage.UserData{
			Subject: session.ClientId,
		}
		globals.OsinServer.FinishAuthorizeRequest(resp, r, ar)
	} else {
		logrus.Debug("Authorization request object is nil")
	}
	logrus.WithField("osin_error", resp.InternalError).Debug("Finished authorize request")
	osin.OutputJSON(resp, w, r)
}
