package interceptors

import "context"

func OpenAIStyleAPIKeyFromContext(ctx context.Context) (string, error) {
	authorization, err := AuthorizationFromContext(ctx)
	if err != nil {
		return "", err
	}

	return BearerFromAuthorization(authorization)
}
