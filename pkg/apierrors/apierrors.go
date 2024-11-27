package apierrors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/bufbuild/protovalidate-go"
	"github.com/labstack/echo/v4"
	"github.com/lingticio/llmg/apis/jsonapi"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
)

type Error struct {
	*jsonapi.ErrorObject

	caller *jsonapi.ErrorCaller

	grpcStatus uint64
	rawError   error
}

func (e *Error) AsStatus() error {
	newStatus := status.New(codes.Code(e.grpcStatus), lo.Ternary(e.Detail == "", e.Title, e.Detail)) //nolint:gosec

	details := []protoiface.MessageV1{e.ErrorObject}
	if e.Caller() != nil {
		details = append(details, e.Caller())
	}

	newStatus, _ = newStatus.WithDetails(details...)

	return newStatus.Err()
}

func (e *Error) AsResponse() *ErrResponse {
	return NewErrResponse().WithError(e)
}

func (e *Error) AsEchoResponse(c echo.Context) error {
	resp := e.AsResponse()
	return c.JSON(resp.HTTPStatus(), resp)
}

func (e *Error) Caller() *jsonapi.ErrorCaller {
	return e.caller
}

func NewError[S ~int, GS ~uint32](status S, grpcStatus GS, code string) *Error {
	return &Error{
		ErrorObject: &jsonapi.ErrorObject{
			Id:     code,
			Status: uint64(status),
			Code:   code,
		},
		grpcStatus: uint64(grpcStatus),
	}
}

func (e *Error) WithError(err error) *Error {
	e.rawError = err
	e.Detail = err.Error()

	return e
}

func (e *Error) WithValidationError(err error) *Error {
	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return e.WithDetail(err.Error())
	}

	validationErrProto := validationErr.ToProto()
	if len(validationErrProto.GetViolations()) == 0 {
		return e.WithDetail(err.Error())
	}

	fieldPath := validationErrProto.GetViolations()[0].GetFieldPath()
	forKey := validationErrProto.GetViolations()[0].GetForKey()
	message := validationErrProto.GetViolations()[0].GetMessage()

	if forKey {
		e.WithDetail(message).WithSourceParameter(fieldPath)
	} else {
		e.WithDetail(message).WithSourcePointer(fieldPath)
	}

	return e
}

func (e *Error) WithCaller() *Error {
	pc, file, line, _ := runtime.Caller(1)

	e.caller = &jsonapi.ErrorCaller{
		Function: runtime.FuncForPC(pc).Name(),
		File:     file,
		Line:     int64(line),
	}

	return e
}

func (e *Error) WithTitle(title string) *Error {
	e.Title = title

	return e
}

func (e *Error) WithDetail(detail string) *Error {
	e.Detail = detail

	return e
}

func (e *Error) WithDetailf(format string, args ...interface{}) *Error {
	e.Detail = fmt.Sprintf(format, args...)

	return e
}

func (e *Error) WithSourcePointer(pointer string) *Error {
	e.Source = &jsonapi.ErrorObjectSource{
		Pointer: pointer,
	}

	return e
}

func (e *Error) WithSourceParameter(parameter string) *Error {
	e.Source = &jsonapi.ErrorObjectSource{
		Parameter: parameter,
	}

	return e
}

func (e *Error) WithSourceHeader(header string) *Error {
	e.Source = &jsonapi.ErrorObjectSource{
		Header: header,
	}

	return e
}

type ErrResponse struct {
	jsonapi.Response
}

func NewErrResponseFromErrorObjects(errs ...*jsonapi.ErrorObject) *ErrResponse {
	resp := NewErrResponse()

	for _, err := range errs {
		resp = resp.WithError(&Error{
			ErrorObject: err,
		})
	}

	return resp
}

func NewErrResponseFromErrorObject(err *jsonapi.ErrorObject) *ErrResponse {
	return NewErrResponse().WithError(&Error{
		ErrorObject: err,
	})
}

func NewErrResponse() *ErrResponse {
	return &ErrResponse{
		Response: jsonapi.Response{
			Errors: make([]*jsonapi.ErrorObject, 0),
		},
	}
}

func (e *ErrResponse) WithError(err *Error) *ErrResponse {
	e.Errors = append(e.Errors, err.ErrorObject)

	return e
}

func (e *ErrResponse) WithValidationError(err error) *ErrResponse {
	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return e.WithError(NewErrInvalidArgument().WithError(err))
	}

	validationErrProto := validationErr.ToProto()
	if len(validationErrProto.GetViolations()) == 0 {
		return e.WithError(NewErrInvalidArgument().WithError(err))
	}

	for _, violation := range validationErrProto.GetViolations() {
		fieldPath := violation.GetFieldPath()
		forKey := violation.GetForKey()
		message := violation.GetMessage()

		if forKey {
			e.WithError(NewErrInvalidArgument().WithDetail(message).WithSourceParameter(fieldPath))
		} else {
			e.WithError(NewErrInvalidArgument().WithDetail(message).WithSourcePointer(fieldPath))
		}
	}

	return e
}

func (e *ErrResponse) HTTPStatus() int {
	if len(e.Errors) == 0 {
		return http.StatusOK
	}

	return int(e.Errors[0].GetStatus()) //nolint:gosec
}
