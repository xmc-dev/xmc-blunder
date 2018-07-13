package certs

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/xmc-dev/xmc/auth-srv/globals"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// TODO: handle error
	asn1, err := x509.MarshalPKIXPublicKey(globals.PubKey)
	if err != nil {
		panic(err)
	}
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: asn1,
		},
	)
	w.Write(pemdata)
}
