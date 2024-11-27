package middlewares

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
)

type ContextKey string

const (
	ContextKeyHeaderAuthorizationAPIKey ContextKey = "header-authorization-api-key"
	ContextKeyHeaderXBaseURL            ContextKey = "header-x-base-url"
)

func HeaderAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("X-Api-Key")
		if apiKey == "" {
			auth := c.Request().Header.Get("Authorization")
			apiKey = strings.TrimPrefix(auth, "Bearer ")
		}

		c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), ContextKeyHeaderAuthorizationAPIKey, apiKey)))

		return next(c)
	}
}

func APIKeyFromContext(ctx context.Context) string {
	apiKey, _ := ctx.Value(ContextKeyHeaderAuthorizationAPIKey).(string)
	return apiKey
}

func HeaderXBaseURL(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		baseURL := c.Request().Header.Get("X-Base-Url")

		c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), ContextKeyHeaderXBaseURL, baseURL)))

		return next(c)
	}
}

func XBaseURLFromContext(ctx context.Context) string {
	baseURL, _ := ctx.Value(ContextKeyHeaderXBaseURL).(string)
	return baseURL
}
