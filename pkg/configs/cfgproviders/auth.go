package authstorage

import (
	"context"

	"github.com/lingticio/llmg/pkg/types/metadata"
)

type Endpoint struct {
	metadata.UnimplementedMetadata

	Tenant   metadata.Tenant                    `json:"tenant" yaml:"tenant"`
	Team     metadata.Team                      `json:"team" yaml:"team"`
	Group    metadata.Group                     `json:"group" yaml:"group"`
	Upstream *metadata.UpstreamSingleOrMultiple `json:"upstream,omitempty" yaml:"upstream,omitempty"`

	ID     string `json:"id" yaml:"id"`
	Alias  string `json:"alias" yaml:"alias"`
	APIKey string `json:"apiKey"`
}

type EndpointProviderQueryable interface {
	FindOneByAPIKey(ctx context.Context, apiKey string) (*Endpoint, error)
	FindOneByAlias(ctx context.Context, alias string) (*Endpoint, error)
}

type EndpointProviderMutable interface {
	ConfigureOneUpstreamForTenant(ctx context.Context, tenantID string, upstream *metadata.UpstreamSingleOrMultiple) error
	ConfigureOneUpstreamForTeam(ctx context.Context, teamID string, upstream *metadata.UpstreamSingleOrMultiple) error
	ConfigureOneUpstreamForGroup(ctx context.Context, groupID string, upstream *metadata.UpstreamSingleOrMultiple) error
	ConfigureOneUpstreamForEndpoint(ctx context.Context, endpointID string, upstream *metadata.UpstreamSingleOrMultiple) error
	ConfigureOne(ctx context.Context, apiKey string, alias string, endpoint *Endpoint) error
}

type EndpointProvider interface {
	EndpointProviderQueryable
}
