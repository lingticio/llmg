package authstorage

import (
	"context"
	"testing"

	"github.com/lingticio/llmg/pkg/types/metadata"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisEndpointProvider_FindOneByAPIKey(t *testing.T) {
	r, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"localhost:6379"},
	})
	require.NoError(t, err)
	require.NotNil(t, r)

	defer r.Close()

	rp := NewRedisEndpointAuthProvider()(r)
	redisProvider, ok := rp.(*RedisEndpointProvider)
	require.True(t, ok)
	require.NotNil(t, redisProvider)

	err = redisProvider.ConfigureOneUpstreamForEndpoint(context.Background(), "endpointId", &metadata.UpstreamSingleOrMultiple{
		Upstream: &metadata.Upstream{
			OpenAI: metadata.UpstreamOpenAI{
				BaseURL: "baseURL",
				APIKey:  "apiKey",
			},
		},
	})
	require.NoError(t, err)

	err = redisProvider.ConfigureOne(context.Background(), "apiKey", "alias", &Endpoint{
		Tenant: metadata.Tenant{Id: "tenantId"},
		Team:   metadata.Team{Id: "teamId"},
		Group:  metadata.Group{Id: "groupId"},
		ID:     "endpointId",
		Alias:  "alias",
		APIKey: "apiKey",
	})
	require.NoError(t, err)

	endpoint, err := rp.FindOneByAPIKey(context.Background(), "apiKey")
	require.NoError(t, err)

	assert.Equal(t, "endpointId", endpoint.ID)
	assert.Equal(t, "alias", endpoint.Alias)
	assert.Equal(t, "apiKey", endpoint.APIKey)
	assert.Equal(t, "tenantId", endpoint.Tenant.Id)
	assert.Equal(t, "teamId", endpoint.Team.Id)
	assert.Equal(t, "groupId", endpoint.Group.Id)
	assert.Equal(t, "baseURL", endpoint.Upstream.OpenAI.BaseURL)
	assert.Equal(t, "apiKey", endpoint.Upstream.OpenAI.APIKey)
}

func TestRedisEndpointProvider_FindOneByAlias(t *testing.T) {
	r, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"localhost:6379"},
	})
	require.NoError(t, err)
	require.NotNil(t, r)

	defer r.Close()

	rp := NewRedisEndpointAuthProvider()(r)
	redisProvider, ok := rp.(*RedisEndpointProvider)
	require.True(t, ok)
	require.NotNil(t, redisProvider)

	err = redisProvider.ConfigureOneUpstreamForEndpoint(context.Background(), "endpointId", &metadata.UpstreamSingleOrMultiple{
		Upstream: &metadata.Upstream{
			OpenAI: metadata.UpstreamOpenAI{
				BaseURL: "baseURL",
				APIKey:  "apiKey",
			},
		},
	})
	require.NoError(t, err)

	err = redisProvider.ConfigureOne(context.Background(), "apiKey", "alias", &Endpoint{
		Tenant: metadata.Tenant{Id: "tenantId"},
		Team:   metadata.Team{Id: "teamId"},
		Group:  metadata.Group{Id: "groupId"},
		ID:     "endpointId",
		Alias:  "alias",
		APIKey: "apiKey",
	})
	require.NoError(t, err)

	endpoint, err := rp.FindOneByAlias(context.Background(), "alias")
	require.NoError(t, err)

	assert.Equal(t, "endpointId", endpoint.ID)
	assert.Equal(t, "alias", endpoint.Alias)
	assert.Equal(t, "apiKey", endpoint.APIKey)
	assert.Equal(t, "tenantId", endpoint.Tenant.Id)
	assert.Equal(t, "teamId", endpoint.Team.Id)
	assert.Equal(t, "groupId", endpoint.Group.Id)
	assert.Equal(t, "baseURL", endpoint.Upstream.OpenAI.BaseURL)
	assert.Equal(t, "apiKey", endpoint.Upstream.OpenAI.APIKey)
}
