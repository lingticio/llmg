package interceptors

import (
	"context"
	"errors"
	"net/http"

	"github.com/lingticio/llmg/pkg/apierrors"
	"google.golang.org/grpc/metadata"
)

func MetadataAuthorization() func(context.Context, *http.Request) metadata.MD {
	return func(ctx context.Context, r *http.Request) metadata.MD {
		md := metadata.MD{}

		authorization := r.Header.Get("Authorization")
		if authorization != "" {
			md.Append("header-authorization", authorization)
		}

		return md
	}
}

func AuthorizationFromMetadata(md metadata.MD) (string, error) {
	values := md.Get("header-authorization")
	if len(values) == 0 {
		return "", nil
	}

	return values[0], nil
}

func AuthorizationFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", apierrors.NewErrInternal().WithError(errors.New("failed to get metadata from context")).WithCaller().AsStatus()
	}

	return AuthorizationFromMetadata(md)
}

func BearerFromAuthorization(authorization string) (string, error) {
	if authorization == "" {
		return "", nil
	}

	if len(authorization) < 7 || authorization[:7] != "Bearer " {
		return "", apierrors.NewBadRequest().WithDetail("malformed Authorization header, expected Bearer token").AsStatus()
	}

	return authorization[7:], nil
}
