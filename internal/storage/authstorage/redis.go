package authstorage

import "context"

var _ EndpointAuthStorage = (*RedisEndpointAuthStorage)(nil)

type RedisEndpointAuthStorage struct {
}

func (s *RedisEndpointAuthStorage) FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}

func (s *RedisEndpointAuthStorage) FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}
