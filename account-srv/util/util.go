package util

import (
	"github.com/xmc-dev/xmc/account-srv/consts"
	"golang.org/x/crypto/bcrypt"
)

func HashSecret(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), consts.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}
