package authstorage

import (
	"context"
	"fmt"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/pkg/types/metadata"
)

var _ EndpointAuthProvider = (*ConfigEndpointAuthProvider)(nil)

type ConfigEndpointAuthProvider struct {
	Config *configs.Routes
}

func (s *ConfigEndpointAuthProvider) findUpstream(endpoint configs.Endpoint, group configs.Group, team configs.Team, tenant configs.Tenant) metadata.Upstreamable {
	if endpoint.Upstream != nil {
		return endpoint.Upstream
	}
	if group.Upstream != nil {
		return group.Upstream
	}
	if team.Upstream != nil {
		return team.Upstream
	}

	return tenant.Upstream
}

func (s *ConfigEndpointAuthProvider) searchGroupsForAPIKey(tenantID, teamID string, groups []configs.Group, apiKey string, team configs.Team, tenant configs.Tenant) (*EndpointAuth, error) {
	for _, group := range groups {
		// Search in current group's endpoints
		for _, endpoint := range group.Endpoints {
			if endpoint.APIKey == apiKey {
				return &EndpointAuth{
					Tenant:   metadata.TenantFromID(tenantID),
					Team:     metadata.TeamFromID(teamID),
					Group:    metadata.GroupFromID(group.ID),
					ID:       endpoint.ID,
					Alias:    endpoint.Alias,
					APIKey:   endpoint.APIKey,
					Upstream: s.findUpstream(endpoint, group, team, tenant),
				}, nil
			}
		}

		// Recursively search in nested groups
		if len(group.Groups) > 0 {
			metadata, err := s.searchGroupsForAPIKey(tenantID, teamID, group.Groups, apiKey, team, tenant)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("api key not found")
}

func (s *ConfigEndpointAuthProvider) FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error) {
	for _, tenant := range s.Config.Tenants {
		for _, team := range tenant.Teams {
			metadata, err := s.searchGroupsForAPIKey(tenant.ID, team.ID, team.Groups, apiKey, team, tenant)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("api key not found")
}

func (s *ConfigEndpointAuthProvider) searchGroupsForAlias(tenantID, teamID string, groups []configs.Group, alias string, team configs.Team, tenant configs.Tenant) (*EndpointAuth, error) {
	for _, group := range groups {
		// Search in current group's endpoints
		for _, endpoint := range group.Endpoints {
			if endpoint.Alias == alias {
				return &EndpointAuth{
					Tenant:   metadata.TenantFromID(tenantID),
					Team:     metadata.TeamFromID(teamID),
					Group:    metadata.GroupFromID(group.ID),
					ID:       endpoint.ID,
					Alias:    endpoint.Alias,
					APIKey:   endpoint.APIKey,
					Upstream: s.findUpstream(endpoint, group, team, tenant),
				}, nil
			}
		}

		// Recursively search in nested groups
		if len(group.Groups) > 0 {
			metadata, err := s.searchGroupsForAlias(tenantID, teamID, group.Groups, alias, team, tenant)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("alias not found")
}

func (s *ConfigEndpointAuthProvider) FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error) {
	for _, tenant := range s.Config.Tenants {
		for _, team := range tenant.Teams {
			metadata, err := s.searchGroupsForAlias(tenant.ID, team.ID, team.Groups, alias, team, tenant)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("alias not found")
}
