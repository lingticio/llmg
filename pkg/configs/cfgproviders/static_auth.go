package authstorage

import (
	"context"
	"fmt"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/pkg/types/metadata"
)

var _ EndpointProvider = (*ConfigEndpointProvider)(nil)

type ConfigEndpointProvider struct {
	Config *configs.Routes
}

func (s *ConfigEndpointProvider) findUpstream(endpoint configs.Endpoint, group configs.Group, team configs.Team, tenant configs.Tenant) *metadata.UpstreamSingleOrMultiple {
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

func (s *ConfigEndpointProvider) searchGroupsForAPIKey(tenantID, teamID string, groups []configs.Group, apiKey string, team configs.Team, tenant configs.Tenant) (*Endpoint, error) {
	for _, group := range groups {
		// Search in current group's endpoints
		for _, endpoint := range group.Endpoints {
			if endpoint.APIKey == apiKey {
				return &Endpoint{
					Tenant:   metadata.Tenant{Id: tenantID},
					Team:     metadata.Team{Id: teamID},
					Group:    metadata.Group{Id: group.ID},
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

func (s *ConfigEndpointProvider) FindOneByAPIKey(ctx context.Context, apiKey string) (*Endpoint, error) {
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

func (s *ConfigEndpointProvider) searchGroupsForAlias(tenantID, teamID string, groups []configs.Group, alias string, team configs.Team, tenant configs.Tenant) (*Endpoint, error) {
	for _, group := range groups {
		// Search in current group's endpoints
		for _, endpoint := range group.Endpoints {
			if endpoint.Alias == alias {
				return &Endpoint{
					Tenant:   metadata.Tenant{Id: tenantID},
					Team:     metadata.Team{Id: teamID},
					Group:    metadata.Group{Id: group.ID},
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

func (s *ConfigEndpointProvider) FindOneByAlias(ctx context.Context, alias string) (*Endpoint, error) {
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
