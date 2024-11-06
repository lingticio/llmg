package authstorage

import (
	"context"
	"fmt"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/pkg/types/metadata"
)

var _ EndpointAuthStorage = (*ConfigEndpointAuthStorage)(nil)

type ConfigEndpointAuthStorage struct {
	Config *configs.Configs
}

func (s *ConfigEndpointAuthStorage) FindMetadataByAPIKey(ctx context.Context, apiKey string) (*EndpointAuth, error) {
	for _, tenant := range s.Config.Tenants {
		for _, team := range tenant.Teams {
			metadata, err := s.searchGroupsForAPIKey(tenant.ID, team.ID, team.Groups, apiKey)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("api key not found")
}

func (s *ConfigEndpointAuthStorage) searchGroupsForAPIKey(tenantID, teamID string, groups []configs.Group, apiKey string) (*EndpointAuth, error) {
	for _, group := range groups {
		// Search in current group's endpoints
		for _, endpoint := range group.Endpoints {
			if endpoint.APIKey == apiKey {
				return &EndpointAuth{
					Tenant: metadata.TenantFromID(tenantID),
					Team:   metadata.TeamFromID(teamID),
					Group:  metadata.GroupFromID(group.ID),
					ID:     endpoint.ID,
					Alias:  endpoint.Alias,
					APIKey: endpoint.APIKey,
				}, nil
			}
		}

		// Recursively search in nested groups
		if len(group.Groups) > 0 {
			metadata, err := s.searchGroupsForAPIKey(tenantID, teamID, group.Groups, apiKey)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("api key not found")
}

func (s *ConfigEndpointAuthStorage) FindMetadataByAlias(ctx context.Context, alias string) (*EndpointAuth, error) {
	for _, tenant := range s.Config.Tenants {
		for _, team := range tenant.Teams {
			metadata, err := s.searchGroupsForAlias(tenant.ID, team.ID, team.Groups, alias)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("alias not found")
}

func (s *ConfigEndpointAuthStorage) searchGroupsForAlias(tenantID, teamID string, groups []configs.Group, alias string) (*EndpointAuth, error) {
	for _, group := range groups {
		// Search in current group's endpoints
		for _, endpoint := range group.Endpoints {
			if endpoint.Alias == alias {
				return &EndpointAuth{
					Tenant: metadata.TenantFromID(tenantID),
					Team:   metadata.TeamFromID(teamID),
					Group:  metadata.GroupFromID(group.ID),
					ID:     endpoint.ID,
					Alias:  endpoint.Alias,
					APIKey: endpoint.APIKey,
				}, nil
			}
		}

		// Recursively search in nested groups
		if len(group.Groups) > 0 {
			metadata, err := s.searchGroupsForAlias(tenantID, teamID, group.Groups, alias)
			if err == nil {
				return metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("alias not found")
}
