package authstorage

import (
	"context"

	"github.com/lingticio/llmg/pkg/types/metadata"
)

type EndpointAuth struct {
	metadata.UnimplementedMetadata

	Tenant metadata.Tenant
	Team   metadata.Team
	Group  metadata.Group

	ID     string
	Alias  string
	APIKey string
}

type EndpointAuthStorage interface {
	FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error)
	FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error)
}
