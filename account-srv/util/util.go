package util

import (
	"golang.org/x/crypto/bcrypt"
	"github.com/xmc-dev/xmc/account-srv/consts"
)

func HashSecret(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), consts.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

