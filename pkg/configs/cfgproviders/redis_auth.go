package authstorage

import (
	"context"
	"encoding/json"

	"github.com/lingticio/llmg/pkg/types/metadata"
	"github.com/lingticio/llmg/pkg/types/redis/rediskeys"
	"github.com/redis/rueidis"
)

var _ EndpointProvider = (*RedisEndpointProvider)(nil)

type RedisEndpointProvider struct {
	rueidis rueidis.Client
}

func NewRedisEndpointAuthProvider() func(rueidis.Client) EndpointProvider {
	return func(r rueidis.Client) EndpointProvider {
		return &RedisEndpointProvider{
			rueidis: r,
		}
	}
}

func (s *RedisEndpointProvider) findUpstreamByRoutesOrGroupID(ctx context.Context, key string) (*metadata.UpstreamSingleOrMultiple, error) {
	cmd := s.rueidis.B().
		Get().
		Key(key).
		Build()

	res, err := s.rueidis.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil
		}

		return nil, err
	}

	var upstream metadata.UpstreamSingleOrMultiple

	err = json.Unmarshal([]byte(res), &upstream)
	if err != nil {
		return nil, err
	}

	return &upstream, nil
}

func (s *RedisEndpointProvider) findUpstreamFromEndpointMetadata(ctx context.Context, endpointMetadata Endpoint) (*metadata.UpstreamSingleOrMultiple, error) {
	pipes := []func(ctx context.Context) (*metadata.UpstreamSingleOrMultiple, error){
		func(ctx context.Context) (*metadata.UpstreamSingleOrMultiple, error) {
			if endpointMetadata.ID == "" {
				return nil, nil
			}

			return s.findUpstreamByRoutesOrGroupID(ctx, rediskeys.EndpointUpstreamByEndpointID1.Format(endpointMetadata.ID))
		},
		func(ctx context.Context) (*metadata.UpstreamSingleOrMultiple, error) {
			if endpointMetadata.Group.ID() == "" {
				return nil, nil
			}

			return s.findUpstreamByRoutesOrGroupID(ctx, rediskeys.EndpointUpstreamByGroupID1.Format(endpointMetadata.Group.ID()))
		},
		func(ctx context.Context) (*metadata.UpstreamSingleOrMultiple, error) {
			if endpointMetadata.Team.ID() == "" {
				return nil, nil
			}

			return s.findUpstreamByRoutesOrGroupID(ctx, rediskeys.EndpointUpstreamByTeamID1.Format(endpointMetadata.Team.ID()))
		},
		func(ctx context.Context) (*metadata.UpstreamSingleOrMultiple, error) {
			if endpointMetadata.Tenant.ID() == "" {
				return nil, nil
			}

			return s.findUpstreamByRoutesOrGroupID(ctx, rediskeys.EndpointUpstreamByTenantID1.Format(endpointMetadata.Tenant.ID()))
		},
	}

	var upstream *metadata.UpstreamSingleOrMultiple

	for _, pipe := range pipes {
		up, err := pipe(ctx)
		if err != nil {
			return nil, err
		}
		if up != nil {
			upstream = up
			break
		}
	}

	if upstream == nil {
		return nil, nil
	}

	return upstream, nil
}

func (s *RedisEndpointProvider) ConfigureOneUpstreamForTenant(ctx context.Context, tenantID string, upstream *metadata.UpstreamSingleOrMultiple) error {
	endpointUpstreamBytes, err := json.Marshal(upstream)
	if err != nil {
		return err
	}

	cmd := s.rueidis.B().
		Set().
		Key(rediskeys.EndpointUpstreamByTenantID1.Format(tenantID)).
		Value(string(endpointUpstreamBytes)).
		Build()

	err = s.rueidis.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisEndpointProvider) ConfigureOneUpstreamForTeam(ctx context.Context, teamID string, upstream *metadata.UpstreamSingleOrMultiple) error {
	endpointUpstreamBytes, err := json.Marshal(upstream)
	if err != nil {
		return err
	}

	cmd := s.rueidis.B().
		Set().
		Key(rediskeys.EndpointUpstreamByTeamID1.Format(teamID)).
		Value(string(endpointUpstreamBytes)).
		Build()

	err = s.rueidis.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisEndpointProvider) ConfigureOneUpstreamForGroup(ctx context.Context, groupID string, upstream *metadata.UpstreamSingleOrMultiple) error {
	endpointUpstreamBytes, err := json.Marshal(upstream)
	if err != nil {
		return err
	}

	cmd := s.rueidis.B().
		Set().
		Key(rediskeys.EndpointUpstreamByGroupID1.Format(groupID)).
		Value(string(endpointUpstreamBytes)).
		Build()

	err = s.rueidis.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisEndpointProvider) ConfigureOneUpstreamForEndpoint(ctx context.Context, endpointID string, upstream *metadata.UpstreamSingleOrMultiple) error {
	endpointUpstreamBytes, err := json.Marshal(upstream)
	if err != nil {
		return err
	}

	cmd := s.rueidis.B().
		Set().
		Key(rediskeys.EndpointUpstreamByEndpointID1.Format(endpointID)).
		Value(string(endpointUpstreamBytes)).
		Build()

	err = s.rueidis.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisEndpointProvider) ConfigureOne(ctx context.Context, apiKey string, alias string, endpoint *Endpoint) error {
	endpointMetadataBytes, err := json.Marshal(endpoint)
	if err != nil {
		return err
	}

	cmd := s.rueidis.B().
		Set().
		Key(rediskeys.EndpointMetadataByAPIKey1.Format(apiKey)).
		Value(string(endpointMetadataBytes)).
		Build()

	err = s.rueidis.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}

	cmd = s.rueidis.B().
		Set().
		Key(rediskeys.EndpointMetadataByAlias1.Format(alias)).
		Value(string(endpointMetadataBytes)).
		Build()

	err = s.rueidis.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisEndpointProvider) FindOneByAPIKey(ctx context.Context, apiKey string) (*Endpoint, error) {
	cmd := s.rueidis.B().
		Get().
		Key(rediskeys.EndpointMetadataByAPIKey1.Format(apiKey)).
		Build()

	res, err := s.rueidis.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil
		}

		return nil, err
	}

	var endpointMetadata Endpoint

	err = json.Unmarshal([]byte(res), &endpointMetadata)
	if err != nil {
		return nil, err
	}

	upstream, err := s.findUpstreamFromEndpointMetadata(ctx, endpointMetadata)
	if err != nil {
		return nil, err
	}
	if upstream == nil {
		return nil, nil
	}

	return &Endpoint{
		Tenant:   endpointMetadata.Tenant,
		Team:     endpointMetadata.Team,
		Group:    endpointMetadata.Group,
		Upstream: upstream,
		ID:       endpointMetadata.ID,
		Alias:    endpointMetadata.Alias,
		APIKey:   endpointMetadata.APIKey,
	}, nil
}

func (s *RedisEndpointProvider) FindOneByAlias(ctx context.Context, alias string) (*Endpoint, error) {
	cmd := s.rueidis.B().
		Get().
		Key(rediskeys.EndpointMetadataByAlias1.Format(alias)).
		Build()

	res, err := s.rueidis.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil
		}

		return nil, err
	}

	var endpointMetadata Endpoint

	err = json.Unmarshal([]byte(res), &endpointMetadata)
	if err != nil {
		return nil, err
	}

	upstream, err := s.findUpstreamFromEndpointMetadata(ctx, endpointMetadata)
	if err != nil {
		return nil, err
	}
	if upstream == nil {
		return nil, nil
	}

	return &Endpoint{
		Tenant:   endpointMetadata.Tenant,
		Team:     endpointMetadata.Team,
		Group:    endpointMetadata.Group,
		Upstream: upstream,
		ID:       endpointMetadata.ID,
		Alias:    endpointMetadata.Alias,
		APIKey:   endpointMetadata.APIKey,
	}, nil
}
