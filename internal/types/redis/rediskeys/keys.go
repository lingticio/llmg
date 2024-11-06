package rediskeys

import "fmt"

// Key key.
type Key string

// Format format.
func (k Key) Format(params ...interface{}) string {
	return fmt.Sprintf(string(k), params...)
}

// Auth.
const (
	// AuthMetadataWithAPIKey key.
	// Params: API key.
	AuthMetadataWithAPIKey1 Key = "auth:credentials:api_key:%s:metadata"
)
