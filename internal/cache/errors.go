package cache

import "errors"

var (
	ErrCacheMiss    = errors.New("cache miss")
	ErrCacheExpired = errors.New("cache expired")
)
