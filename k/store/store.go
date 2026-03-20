package store

import (
	"github.com/mojocn/base64Captcha"
)

type cacheStore struct {
	cache      AdapterCache
	expiration int
}

func NewCacheStore(cache AdapterCache, expiration int) base64Captcha.Store {
	s := new(cacheStore)
	s.cache = cache
	s.expiration = expiration
	return s
}

func (e *cacheStore) Set(id string, value string) error {
	return e.cache.Set(id, value, e.expiration)
}

func (e *cacheStore) Get(id string, clear bool) string {
	v, err := e.cache.Get(id)
	if err == nil {
		if clear {
			_ = e.cache.Del(id)
		}
		return v
	}
	return ""
}

func (e *cacheStore) Verify(id, answer string, clear bool) bool {
	return e.Get(id, clear) == answer
}
