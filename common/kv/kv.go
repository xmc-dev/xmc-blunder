package kv

import (
	"errors"

	"github.com/hashicorp/consul/api"
)

var ErrNoKey = errors.New("no such key in consul")

func Get(kv *api.KV, key string) (*api.KVPair, error) {
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, ErrNoKey
	}

	return pair, nil
}
