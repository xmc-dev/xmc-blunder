// Package config provides utility functions for reading, exposing and manipulating
// service configuration values stored in Consul.
package config

import (
	"os"
	"path"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

func mustSet(key, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		panic(err)
	}
}

// Reader reads and manipulates config values from consul
type Reader struct {
	client   *consul.Client
	prefixes []string
}

// NewReader creates a new reader
func NewReader(client *consul.Client, prefixes ...string) *Reader {
	cr := &Reader{client: client, prefixes: prefixes}

	return cr
}

func envName(base, p string) string {
	ret := []string{"CFG"}

	tail := strings.TrimPrefix(base, p+"/")
	ret = append(ret, strings.Split(tail, "/")...)

	return strings.ToUpper(strings.Join(ret, "_"))
}

// ReadConfig reads the config values and exposes them as environment variables.
// The env var's name is based on the config key's name without the given prefix.
//
// Example: xmc/core/config/db/user with prefix xmc/core/config -> CFG_DB_USER
func (cr *Reader) ReadConfig() error {
	for _, p := range cr.prefixes {
		p = path.Clean(p)
		pairs, _, err := cr.client.KV().List(p, nil)
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			if pair.Flags != 1 {
				continue
			}
			mustSet(envName(pair.Key, p), string(pair.Value))
		}
	}

	return nil
}

// MustReadConfig calls ReadConfig and panics on error
func (cr *Reader) MustReadConfig() {
	err := cr.ReadConfig()
	if err != nil {
		panic(err)
	}
}
