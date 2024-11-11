package rediskeys

import "fmt"

// Key key.
type Key string

// Format format.
func (k Key) Format(params ...interface{}) string {
	return fmt.Sprintf(string(k), params...)
}

// Endpoint Provider

const (
	// EndpointMetadataByAPIKey1.
	// Params: API Key.
	EndpointMetadataByAPIKey1 Key = "config:providers:auth:metadata:api_key:%s"

	// EndpointUpstreamByTenantID1.
	// Params: Tenant ID.
	EndpointUpstreamByTenantID1 Key = "config:providers:auth:metadata:upstream:tenant:%s"

	// EndpointUpstreamByTeamID1.
	// Params: Team ID.
	EndpointUpstreamByTeamID1 Key = "config:providers:auth:metadata:upstream:team:%s"

	// EndpointUpstreamByGroupID1.
	// Params: Group ID.
	EndpointUpstreamByGroupID1 Key = "config:providers:auth:metadata:upstream:group:%s"

	// EndpointUpstreamByEndpointID1.
	// Params: Endpoint ID.
	EndpointUpstreamByEndpointID1 Key = "config:providers:auth:metadata:upstream:endpoint:%s"

	// EndpointMetadataByAlias1.
	// Params: Alias.
	EndpointMetadataByAlias1 Key = "config:providers:auth:metadata:alias:%s"
)
