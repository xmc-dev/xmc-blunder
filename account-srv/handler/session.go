package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/account-srv/consts"
	"github.com/xmc-dev/xmc/account-srv/db"
	maccount "github.com/xmc-dev/xmc/account-srv/db/models/account"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/account-srv/proto/session"
	"golang.org/x/crypto/bcrypt"
)

type SessionsService struct{}

func sessSName(method string) string {
	return fmt.Sprintf("%s.SessionsService.%s", consts.ServiceName, method)
}

func (*SessionsService) Login(_ context.Context, req *session.LoginRequest, rsp *session.LoginResponse) error {
	methodName := sessSName("Login")
	req.ClientId = strings.ToLower(req.ClientId)

	switch {
	case len(req.ClientId) == 0:
		return errors.BadRequest(methodName, "client_id cannot be blank")
	}

	accs, err := db.SearchAccount(&account.SearchRequest{
		ClientId: req.ClientId,
		Type:     req.Type,
	})

	if err != nil || len(accs) == 0 {
		if len(accs) == 0 {
			return errors.NotFound(methodName, "account not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	if len(accs) > 1 {
		panic("got more than one response")
	}
	acc := accs[0]

	if req.Type != nil && acc.Type != maccount.Type(req.Type.Value) {
		return errors.NotFound(methodName, "account not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.ClientSecret), []byte(req.ClientSecret))
	logrus.WithFields(logrus.Fields{
		"acc_secret": string([]byte(acc.ClientSecret)),
		"req_secret": req.ClientSecret,
	}).Debug("Comparing passwords")
	if err != nil {
		logrus.WithField("err", err).Debug("Password comparison failed")
		return errors.Unauthorized(methodName, "client_secret incorrect")
	}

	if acc.Type == maccount.USER {
		sess := &session.Session{
			ClientId:  req.ClientId,
			CreatedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(24 * time.Hour * 7).Unix(),
		}
		ss, err := db.CreateSession(sess)
		if err != nil {
			return errors.InternalServerError(methodName, err.Error())
		}

		rsp.Session = ss.ToProto()
	}
	rsp.CredentialsOk = true

	return nil
}

func (*SessionsService) Read(_ context.Context, req *session.ReadRequest, rsp *session.ReadResponse) error {
	methodName := sessSName("Read")
	if len(req.Uuid) == 0 {
		return errors.BadRequest(methodName, "uuid cannot be blank")
	}

	uuid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return errors.BadRequest(methodName, "uuid cannot be blank")
	}
	sess, err := db.ReadSession(uuid)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "session not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	rsp.Session = sess.ToProto()

	return nil
}

func (*SessionsService) Logout(_ context.Context, req *session.LogoutRequest, _ *session.LogoutResponse) error {
	methodName := sessSName("Logout")
	if len(req.Uuid) == 0 {
		return errors.BadRequest(methodName, "uuid cannot be blank")
	}

	uuid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return errors.BadRequest(methodName, "uuid cannot be blank")
	}
	err = db.DeleteSession(uuid)

	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "session not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	return nil
}
