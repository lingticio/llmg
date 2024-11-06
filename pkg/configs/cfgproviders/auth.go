package authstorage

import (
	"context"

	"github.com/lingticio/llmg/pkg/types/metadata"
)

type EndpointAuth struct {
	metadata.UnimplementedMetadata

	Tenant   metadata.Tenant
	Team     metadata.Team
	Group    metadata.Group
	Upstream metadata.Upstreamable

	ID     string
	Alias  string
	APIKey string
}

type EndpointAuthProvider interface {
	FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error)
	FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error)
}
