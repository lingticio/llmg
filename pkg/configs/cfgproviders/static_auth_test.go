package authstorage

import (
	"context"
	"testing"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/pkg/types/metadata"
	"github.com/nekomeowww/xo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigEndpointProvider_FindOneByAPIKey(t *testing.T) {
	tenantID := xo.RandomHashString(8)
	teamID := xo.RandomHashString(8)
	groupID1 := xo.RandomHashString(8)
	groupID2 := xo.RandomHashString(8)

	apiKey1 := xo.RandomHashString(16)
	apiKey2 := xo.RandomHashString(16)

	s := &ConfigEndpointProvider{
		Config: &configs.Routes{
			Tenants: []configs.Tenant{
				{
					ID: tenantID,
					Teams: []configs.Team{
						{
							ID: teamID,
							Groups: []configs.Group{
								{
									ID: groupID1,
									Groups: []configs.Group{
										{
											ID: groupID2,
											Endpoints: []configs.Endpoint{
												{
													ID:     xo.RandomHashString(8),
													Alias:  "test",
													APIKey: apiKey1,
												},
											},
										},
									},
									Endpoints: []configs.Endpoint{
										{
											ID:     xo.RandomHashString(8),
											Alias:  "test-2",
											APIKey: apiKey2,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	md, err := s.FindOneByAPIKey(context.TODO(), apiKey1)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID2, md.Group.ID())
	assert.Equal(t, apiKey1, md.APIKey)

	md, err = s.FindOneByAPIKey(context.TODO(), apiKey2)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID1, md.Group.ID())
	assert.Equal(t, apiKey2, md.APIKey)

	md, err = s.FindOneByAPIKey(context.TODO(), "invalid")
	require.Error(t, err)
	require.Nil(t, md)
}

func TestConfigEndpointProvider_FindOneByAlias(t *testing.T) {
	tenantID := xo.RandomHashString(8)
	teamID := xo.RandomHashString(8)
	groupID1 := xo.RandomHashString(8)
	groupID2 := xo.RandomHashString(8)

	alias1 := "test"
	alias2 := "test-2"

	s := &ConfigEndpointProvider{
		Config: &configs.Routes{
			Tenants: []configs.Tenant{
				{
					ID: tenantID,
					Teams: []configs.Team{
						{
							ID: teamID,
							Groups: []configs.Group{
								{
									ID: groupID1,
									Groups: []configs.Group{
										{
											ID: groupID2,
											Endpoints: []configs.Endpoint{
												{
													ID:     xo.RandomHashString(8),
													Alias:  alias1,
													APIKey: xo.RandomHashString(16),
												},
											},
										},
									},
									Endpoints: []configs.Endpoint{
										{
											ID:     xo.RandomHashString(8),
											Alias:  alias2,
											APIKey: xo.RandomHashString(16),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	md, err := s.FindOneByAlias(context.TODO(), alias1)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID2, md.Group.ID())
	assert.Equal(t, alias1, md.Alias)

	md, err = s.FindOneByAlias(context.TODO(), alias2)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID1, md.Group.ID())
	assert.Equal(t, alias2, md.Alias)

	md, err = s.FindOneByAlias(context.TODO(), "invalid")
	require.Error(t, err)
	require.Nil(t, md)
}

func TestConfigAuthStorage_findUpstream(t *testing.T) {
	tenantID := xo.RandomHashString(8)
	teamID := xo.RandomHashString(8)
	groupID := xo.RandomHashString(8)

	tenantUpstream := &metadata.UpstreamSingleOrMultiple{
		Upstream: &metadata.Upstream{
			OpenAI: metadata.UpstreamOpenAI{
				BaseURL: "tenant-url",
				APIKey:  "tenant-key",
			},
		},
	}

	teamUpstream := &metadata.UpstreamSingleOrMultiple{
		Upstream: &metadata.Upstream{
			OpenAI: metadata.UpstreamOpenAI{
				BaseURL: "team-url",
				APIKey:  "team-key",
			},
		},
	}

	groupUpstream := &metadata.UpstreamSingleOrMultiple{
		Upstream: &metadata.Upstream{
			OpenAI: metadata.UpstreamOpenAI{
				BaseURL: "group-url",
				APIKey:  "group-key",
			},
		},
	}

	endpointUpstream := &metadata.UpstreamSingleOrMultiple{
		Upstream: &metadata.Upstream{
			OpenAI: metadata.UpstreamOpenAI{
				BaseURL: "endpoint-url",
				APIKey:  "endpoint-key",
			},
		},
	}

	s := &ConfigEndpointProvider{
		Config: &configs.Routes{
			Tenants: []configs.Tenant{
				{
					ID:       tenantID,
					Upstream: tenantUpstream,
					Teams: []configs.Team{
						{
							ID:       teamID,
							Upstream: teamUpstream,
							Groups: []configs.Group{
								{
									ID:       groupID,
									Upstream: groupUpstream,
									Endpoints: []configs.Endpoint{
										{
											ID:       "endpoint1",
											APIKey:   "key1",
											Upstream: endpointUpstream,
										},
										{
											ID:     "endpoint2",
											APIKey: "key2",
											// No upstream - should inherit from group
										},
									},
								},
								{
									ID: "group2",
									Endpoints: []configs.Endpoint{
										{
											ID:     "endpoint3",
											APIKey: "key3",
											// No upstream - should inherit from team
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Test endpoint with its own upstream
	md1, err := s.FindOneByAPIKey(context.TODO(), "key1")
	require.NoError(t, err)
	assert.Equal(t, endpointUpstream, md1.Upstream)

	// Test endpoint inheriting from group
	md2, err := s.FindOneByAPIKey(context.TODO(), "key2")
	require.NoError(t, err)
	assert.Equal(t, groupUpstream, md2.Upstream)

	// Test endpoint inheriting from team
	md3, err := s.FindOneByAPIKey(context.TODO(), "key3")
	require.NoError(t, err)
	assert.Equal(t, teamUpstream, md3.Upstream)
}
