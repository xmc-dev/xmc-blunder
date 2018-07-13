package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/micro/go-micro/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/account-srv/consts"
	"github.com/xmc-dev/xmc/account-srv/db"
	maccount "github.com/xmc-dev/xmc/account-srv/db/models/account"
	"github.com/xmc-dev/xmc/account-srv/proto/account"
	"github.com/xmc-dev/xmc/account-srv/util"
)

const (
	clientIDMaxLen   = 50
	clientIDMinLen   = 3
	clientSecretLen  = 64
	clientNameMaxLen = 50
)

// AccountsService is a service for managing accounts
type AccountsService struct{}

// ACCounts Service Name
func accSName(method string) string {
	return fmt.Sprintf("%s.AccountsService.%s", consts.ServiceName, method)
}

func validateCallbackURL(methodName, callbackURL string) error {
	if len(callbackURL) == 0 {
		return errors.BadRequest(methodName, "callback_url cannot be blank")
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}
	if !u.IsAbs() {
		return errors.BadRequest(methodName, "callback_url is not absolute")
	}

	return nil
}

const genClientIDAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
const genClientSecretAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var genClientIDAlphabetI = big.NewInt(int64(len(genClientIDAlphabet)))
var genClientSecretAlphabetI = big.NewInt(int64(len(genClientSecretAlphabet)))

func genClientID(len int) string {
	b := make([]byte, len)
	for i := range b {
		rnd, _ := rand.Int(rand.Reader, genClientIDAlphabetI)
		b[i] = genClientIDAlphabet[rnd.Uint64()]
	}

	return string(b)
}

func genClientSecret(secretLen int) string {
	b := make([]byte, secretLen)
	for i := range b {
		rnd, _ := rand.Int(rand.Reader, genClientSecretAlphabetI)
		b[i] = genClientSecretAlphabet[rnd.Uint64()]
	}

	return string(b)
}

func isAllGraphic(str string) bool {
	for _, r := range str {
		if !unicode.IsGraphic(r) {
			return false
		}
	}

	return true
}

func isClientIDValid(acc *account.Account) bool {
	if acc.Type == account.Type_SERVICE {
		return true
	}
	return len(acc.ClientId) > 0 && len(acc.ClientId) >= clientIDMinLen && len(acc.ClientId) <= clientIDMaxLen && isAllGraphic(acc.ClientId)
}

func isNameValid(name string) bool {
	return len(name) > 0 && // not nil
		len(name) <= clientNameMaxLen && // has maximum limit
		isAllGraphic(name)
}

func isClientSecretValid(secret string) bool {
	return len(secret) > 0 && len(secret) <= clientSecretLen
}

// Create creates an account
func (acc *AccountsService) Create(ctx context.Context, req *account.CreateRequest, rsp *account.CreateResponse) error {
	methodName := accSName("Create")
	a := req.Account
	a.ClientId = strings.ToLower(a.ClientId)

	a.OwnerUuid = strings.ToLower(a.OwnerUuid)

	a.IsFirstParty = false
	switch {
	case a == nil:
		return errors.BadRequest(methodName, "invalid account")
	case !isClientIDValid(a):
		return errors.BadRequest(methodName, "invalid client_id")
	case !isClientSecretValid(a.ClientSecret):
		return errors.BadRequest(methodName, "invalid client_secret")
	case !isNameValid(a.Name):
		return errors.BadRequest(methodName, "invalid name")
	case len(a.ClientSecret) == 0 && a.Type != account.Type_SERVICE:
		return errors.BadRequest(methodName, "client_secret cannot be blank")
	case a.IsPublic && a.Type == account.Type_USER:
		return errors.BadRequest(methodName, "user accounts can't be public")
	case len(a.Scope) > 0 && a.Type == account.Type_USER:
		return errors.BadRequest(methodName, "user accounts don't have a scope")
	case len(a.RoleId) == 0 && a.Type == account.Type_USER:
		a.RoleId = "default"
	case a.Type == account.Type_SERVICE:
		err := validateCallbackURL(methodName, a.CallbackUrl)
		if err != nil {
			return err
		}
		if len(a.OwnerUuid) == 0 {
			return errors.BadRequest(methodName, "invalid owner_uuid")
		} else {
			// check if owner exists
			ownerUUID, err := uuid.Parse(a.OwnerUuid)
			if err != nil {
				return errors.BadRequest(methodName, "invalid owner_uuid")
			}
			_, dbErr := db.ReadAccount(ownerUUID)
			if dbErr != nil {
				if dbErr == db.ErrNotFound {
					return errors.NotFound(methodName, "owner_uuid not found")
				} else {
					return errors.InternalServerError(methodName, dbErr.Error())
				}
			}
		}
	}

	if a.Type == account.Type_SERVICE {
		// user accounts are not allowed to have a client_id
		// of clientIDMaxLen + 1 length. so the chance of conflict is much smaller
		a.ClientId = genClientID(clientIDMaxLen + 1)
		a.RoleId = ""
		if a.IsPublic {
			a.ClientSecret = ""
		} else {
			a.ClientSecret = genClientSecret(clientSecretLen)
		}
		rsp.ClientId = a.ClientId
		rsp.ClientSecret = a.ClientSecret
	}
	hash, err := util.HashSecret(a.ClientSecret)
	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}

	a.ClientSecret = hash

	id, err := db.CreateAccount(a)
	if err != nil {
		if err == db.ErrUniqueViolation {
			if a.Type == account.Type_SERVICE {
				logrus.WithField("client_id", a.ClientId).Warn("client_id service account conflict")
			}
			return errors.Conflict(methodName, "an account with the same client_id already exists")
		}

		return errors.InternalServerError(methodName, err.Error())
	}
	rsp.Uuid = id.String()

	return nil
}

// Read returns an account
func (acc *AccountsService) Read(ctx context.Context, req *account.ReadRequest, rsp *account.ReadResponse) error {
	methodName := accSName("Read")
	if len(req.Uuid) == 0 {
		return errors.BadRequest(methodName, "uuid cannot be blank")
	}

	uuid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return errors.BadRequest(methodName, "invalid uuid")
	}
	a, err := db.ReadAccount(uuid)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "account not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	rsp.Account = a.ToProto()
	// hide secret
	rsp.Account.ClientSecret = ""
	return nil
}

func (acc *AccountsService) Get(ctx context.Context, req *account.GetRequest, rsp *account.GetResponse) error {
	methodName := accSName("Get")
	if len(req.ClientId) == 0 {
		return errors.BadRequest(methodName, "client_id cannot be blank")
	}
	req.ClientId = strings.ToLower(req.ClientId)

	a, err := db.GetAccount(req.ClientId)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "account not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}
	rsp.Account = a.ToProto()
	rsp.Account.ClientSecret = ""
	return nil
}

// Update updates an account's fields
func (acc *AccountsService) Update(ctx context.Context, req *account.UpdateRequest, rsp *account.UpdateResponse) error {
	methodName := accSName("Update")

	if len(req.Uuid) == 0 {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	uuid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return errors.BadRequest(methodName, "invalid uuid")
	}
	a, err := db.ReadAccount(uuid)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "account not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	if req.ClientSecret != "" {
		if a.Type == maccount.SERVICE {
			return errors.BadRequest(methodName, "can't change secret, account is service")
		}
		hash, err := util.HashSecret(req.ClientSecret)
		if err != nil {
			return errors.InternalServerError(methodName, err.Error())
		}
		req.ClientSecret = hash
	}

	if len(req.CallbackUrl) > 0 {
		if a.Type == maccount.USER {
			return errors.BadRequest(methodName, "users don't have a callback_url")
		}
		if err := validateCallbackURL(methodName, req.CallbackUrl); err != nil {
			return err
		}
	}

	if len(req.Name) > 0 {
		if !isNameValid(req.Name) {
			return errors.BadRequest(methodName, "invalid name")
		}
	}

	err = db.UpdateAccount(req)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "account not found")
		} else if _, ok := err.(db.ErrHasDependants); ok {
			return errors.BadRequest(methodName, "client_id already in use or role_id not found")
		}
		return errors.InternalServerError(methodName, err.Error())
	}

	return nil
}

// Delete deletes an account
func (acc *AccountsService) Delete(ctx context.Context, req *account.DeleteRequest, rsp *account.DeleteResponse) error {
	methodName := accSName("Delete")
	if len(req.Uuid) == 0 {
		return errors.BadRequest(methodName, "invalid uuid")
	}

	errNf := func(err error) error {
		if err == db.ErrNotFound {
			return errors.NotFound(methodName, "account not found")
		} else if err != nil {
			return errors.InternalServerError(methodName, err.Error())
		}
		return nil
	}

	delAccount := func(uuid uuid.UUID) error {
		err := db.DeleteAccount(uuid)
		if err != nil {
			if errNf(err) != nil {
				return errNf(err)
			}
		}
		return nil
	}

	uuid, err := uuid.Parse(req.Uuid)
	if err != nil {
		return errors.BadRequest(methodName, "invalid uuid")
	}
	a, err := db.ReadAccount(uuid)
	if errNf(err) != nil {
		return errNf(err)
	}
	// if user, delete all its service accounts
	if a.Type == maccount.USER {
		services, err := db.SearchAccount(&account.SearchRequest{
			OwnerUuid: a.UUID.String(),
		})
		if err != nil && err != db.ErrNotFound {
			return errors.InternalServerError(methodName, err.Error())
		}
		if err == nil {
			for _, s := range services {
				err := delAccount(s.UUID)
				if err != nil {
					return err
				}
			}
		}
	}
	return delAccount(a.UUID)
}

// Search returns all accounts that match the query
func (acc *AccountsService) Search(ctx context.Context, req *account.SearchRequest, rsp *account.SearchResponse) error {
	methodName := accSName("Search")
	if req.Limit == 0 {
		req.Limit = 10
	}

	accs, err := db.SearchAccount(req)

	if err != nil {
		return errors.InternalServerError(methodName, err.Error())
	}

	rsp.Accounts = []*account.Account{}
	for _, acc := range accs {
		pacc := acc.ToProto()
		pacc.ClientSecret = ""
		rsp.Accounts = append(rsp.Accounts, pacc)
	}

	// hide secret
	for _, racc := range rsp.Accounts {
		racc.ClientSecret = ""
	}

	return nil
}
