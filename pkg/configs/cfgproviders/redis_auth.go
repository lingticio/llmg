package authstorage

import "context"

var _ EndpointAuthProvider = (*RedisEndpointAuthProvider)(nil)

type RedisEndpointAuthProvider struct {
}

func (s *RedisEndpointAuthProvider) FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}

func (s *RedisEndpointAuthProvider) FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}
