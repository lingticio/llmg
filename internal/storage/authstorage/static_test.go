package authstorage

import (
	"context"
	"testing"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/nekomeowww/xo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigEndpointAuthStorage_FindMetadataByAPIKey(t *testing.T) {
	tenantID := xo.RandomHashString(8)
	teamID := xo.RandomHashString(8)
	groupID1 := xo.RandomHashString(8)
	groupID2 := xo.RandomHashString(8)

	apiKey1 := xo.RandomHashString(16)
	apiKey2 := xo.RandomHashString(16)

	s := &ConfigEndpointAuthStorage{
		Config: &configs.Configs{
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

	md, err := s.FindMetadataByAPIKey(context.TODO(), apiKey1)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID2, md.Group.ID())
	assert.Equal(t, apiKey1, md.APIKey)

	md, err = s.FindMetadataByAPIKey(context.TODO(), apiKey2)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID1, md.Group.ID())
	assert.Equal(t, apiKey2, md.APIKey)

	md, err = s.FindMetadataByAPIKey(context.TODO(), "invalid")
	require.Error(t, err)
	require.Nil(t, md)
}

func TestConfigEndpointAuthStorage_FindMetadataByAlias(t *testing.T) {
	tenantID := xo.RandomHashString(8)
	teamID := xo.RandomHashString(8)
	groupID1 := xo.RandomHashString(8)
	groupID2 := xo.RandomHashString(8)

	alias1 := "test"
	alias2 := "test-2"

	s := &ConfigEndpointAuthStorage{
		Config: &configs.Configs{
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

	md, err := s.FindMetadataByAlias(context.TODO(), alias1)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID2, md.Group.ID())
	assert.Equal(t, alias1, md.Alias)

	md, err = s.FindMetadataByAlias(context.TODO(), alias2)
	require.NoError(t, err)
	require.NotNil(t, md)

	assert.Equal(t, tenantID, md.Tenant.ID())
	assert.Equal(t, teamID, md.Team.ID())
	assert.Equal(t, groupID1, md.Group.ID())
	assert.Equal(t, alias2, md.Alias)

	md, err = s.FindMetadataByAlias(context.TODO(), "invalid")
	require.Error(t, err)
	require.Nil(t, md)
}
