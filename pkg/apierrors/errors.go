package apierrors

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

func NewBadRequest() *Error {
	return NewError(http.StatusBadRequest, codes.InvalidArgument, "BAD_REQUEST").
		WithTitle("Bad Request").
		WithDetail("The request was invalid or cannot be served")
}

func NewErrInternal() *Error {
	return NewError(http.StatusInternalServerError, codes.Internal, "INTERNAL_SERVER_ERROR").
		WithTitle("Internal Server Error").
		WithDetail("An internal server error occurred")
}

func NewPermissionDenied() *Error {
	return NewError(http.StatusForbidden, codes.PermissionDenied, "PERMISSION_DENIED").
		WithTitle("Permission Denied").
		WithDetail("You do not have permission to access the requested resources")
}

func NewErrUnavailable() *Error {
	return NewError(http.StatusServiceUnavailable, codes.Unavailable, "UNAVAILABLE").
		WithTitle("Service Unavailable").
		WithDetail("The requested service is unavailable")
}

func NewErrInvalidArgument() *Error {
	return NewError(http.StatusBadRequest, codes.InvalidArgument, "INVALID_ARGUMENT").
		WithTitle("Invalid Argument").
		WithDetail("Invalid parameters, queries, body, or headers were sent, please check the request")
}

func NewErrUnauthorized() *Error {
	return NewError(http.StatusUnauthorized, codes.Unauthenticated, "UNAUTHORIZED").
		WithTitle("Unauthorized").
		WithDetail("The requested resources require authentication")
}

func NewErrNotFound() *Error {
	return NewError(http.StatusNotFound, codes.NotFound, "NOT_FOUND").
		WithTitle("Not Found").
		WithDetail("The requested resources were not found")
}

func NewPaymentRequired() *Error {
	return NewError(http.StatusPaymentRequired, codes.FailedPrecondition, "PAYMENT_REQUIRED").
		WithTitle("Payment Required").
		WithDetail("The requested resources require payment")
}

func NewQuotaExceeded() *Error {
	return NewError(http.StatusTooManyRequests, codes.ResourceExhausted, "QUOTA_EXCEEDED").
		WithTitle("Quota Exceeded").
		WithDetail("The request quota has been exceeded")
}

func NewForbidden() *Error {
	return NewError(http.StatusForbidden, codes.PermissionDenied, "FORBIDDEN").
		WithTitle("Forbidden").
		WithDetail("You do not have permission to access the requested resources")
}
