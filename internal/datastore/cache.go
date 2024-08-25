package datastore

import (
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/fx"
)

type NewCacheParams struct {
	fx.In
}

type Cache struct {
	*cache.Cache
}

func NewCache() func(params NewCacheParams) (*Cache, error) {
	return func(params NewCacheParams) (*Cache, error) {
		return &Cache{
			Cache: cache.New(time.Hour, time.Hour),
		}, nil
	}
}
