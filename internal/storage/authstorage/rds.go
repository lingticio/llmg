package authstorage

import "context"

var _ EndpointAuthStorage = (*RDSEndpointAuthStorage)(nil)

type RDSEndpointAuthStorage struct {
}

func (s *RDSEndpointAuthStorage) FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}

func (s *RDSEndpointAuthStorage) FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}
