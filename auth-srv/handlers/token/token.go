package token

import (
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/auth-srv/globals"
	"github.com/xmc-dev/xmc/auth-srv/storage"
	"github.com/xmc-dev/xmc/auth-srv/util"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	resp := globals.OsinServer.NewResponse()
	defer resp.Close()

	if ar := globals.OsinServer.HandleAccessRequest(resp, r); ar != nil {
		if ar.Type == osin.CLIENT_CREDENTIALS {
			c := ar.Client.(*storage.Client)
			logrus.WithFields(logrus.Fields{"original": c.Original.Scope, "scope": ar.Scope}).Debug("Client credentials requesting scopes")
			ok := util.HasScopes(c.Original.Scope, ar.Scope)
			ar.Authorized = ok
		} else {
			ar.Authorized = true
		}
		globals.OsinServer.FinishAccessRequest(resp, r, ar)
	}
	logrus.WithField("osin_error", resp.InternalError).Debug("Finished token request")
	osin.OutputJSON(resp, w, r)
}
