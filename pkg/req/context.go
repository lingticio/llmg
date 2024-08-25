package req

import (
	"context"
	"net/http"
)

type contextKey int

const (
	requestKey contextKey = iota
)

func WithContextRequest(ctx context.Context, request *http.Request) context.Context {
	return context.WithValue(ctx, requestKey, request)
}

func FromContextRequest(ctx context.Context) *http.Request {
	req, _ := ctx.Value(requestKey).(*http.Request)
	return req
}
