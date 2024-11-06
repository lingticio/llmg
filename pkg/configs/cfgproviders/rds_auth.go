package authstorage

import "context"

var _ EndpointAuthProvider = (*RDSEndpointAuthProvider)(nil)

type RDSEndpointAuthProvider struct {
}

func (s *RDSEndpointAuthProvider) FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}

func (s *RDSEndpointAuthProvider) FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error) {
	return &EndpointAuth{}, nil
}
