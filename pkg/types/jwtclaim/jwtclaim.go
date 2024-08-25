package jwtclaim

import "github.com/golang-jwt/jwt"

var _ jwt.Claims = (*SupabaseAccessTokenClaim)(nil)

type SupabaseAccessTokenClaim struct {
	Aal          string                               `json:"aal"`
	Amr          []SupabaseAccessTokenClaimArm        `json:"amr"`
	AppMetadata  SupabaseAccessTokenClaimAppMetadata  `json:"app_metadata"`
	Aud          any                                  `json:"aud"` // 2024.7.12, string changed to []string
	Email        string                               `json:"email"`
	Exp          int                                  `json:"exp"`
	Iat          int                                  `json:"iat"`
	IsAnonymous  bool                                 `json:"is_anonymous"`
	Iss          string                               `json:"iss"`
	Phone        string                               `json:"phone"`
	Role         string                               `json:"role"`
	SessionID    string                               `json:"session_id"`
	Sub          string                               `json:"sub"`
	UserMetadata SupabaseAccessTokenClaimUserMetadata `json:"user_metadata"`
}

type SupabaseAccessTokenClaimArm struct {
	Method    string `json:"method"`
	Timestamp int    `json:"timestamp"`
}

type SupabaseAccessTokenClaimAppMetadata struct {
	Provider  string   `json:"provider"`
	Providers []string `json:"providers"`
}

type SupabaseAccessTokenClaimUserMetadata struct {
	AvatarURL     string `json:"avatar_url"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FullName      string `json:"full_name"`
	Iss           string `json:"iss"`
	Name          string `json:"name"`
	PhoneVerified bool   `json:"phone_verified"`
	Picture       string `json:"picture"`
	ProviderID    string `json:"provider_id"`
	Sub           string `json:"sub"`
}

func (c *SupabaseAccessTokenClaim) Valid() error {
	return nil
}

type UserProfileClaimUserMetadata struct {
	Type   string `json:"type"`
	Scope  string `json:"scope"`
	UserID string `json:"user_id"`
}

type UserProfileClaim struct {
	*jwt.StandardClaims
	UserMetadata UserProfileClaimUserMetadata `json:"user_metadata"`
}

func (c *UserProfileClaim) Valid() error {
	if c.StandardClaims == nil {
		return jwt.NewValidationError("missing standard claims", jwt.ValidationErrorClaimsInvalid)
	}

	err := c.StandardClaims.Valid()
	if err != nil {
		return err
	}

	if c.UserMetadata.Type == "" || c.UserMetadata.UserID == "" || c.UserMetadata.Scope == "" {
		return jwt.NewValidationError("missing user metadata", jwt.ValidationErrorClaimsInvalid)
	}

	return nil
}
