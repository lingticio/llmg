package authstorage

import "context"

var _ EndpointProvider = (*RDSEndpointAuthProvider)(nil)

type RDSEndpointAuthProvider struct {
}

func (s *RDSEndpointAuthProvider) FindOneByAPIKey(ctx context.Context, apiKey string) (*Endpoint, error) {
	return &Endpoint{}, nil
}

func (s *RDSEndpointAuthProvider) FindOneByAlias(ctx context.Context, alias string) (*Endpoint, error) {
	return &Endpoint{}, nil
}
