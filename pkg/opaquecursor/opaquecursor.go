package opaquecursor

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/samber/lo"
)

type OpaqueCursor[F any, O any] struct {
	Filters  F `json:"filters"`
	OrderBys O `json:"orderBys"`
}

func (oc OpaqueCursor[F, S]) IsValid() bool {
	refValue := reflect.ValueOf(oc.OrderBys)
	if refValue.Kind() == reflect.Struct {
		for i := 0; i < refValue.NumField(); i++ {
			if lo.Contains([]string{"ASC", "DESC"}, strings.ToUpper(refValue.Field(i).String())) {
				return true
			}
		}
	}

	return false
}

func Unmarshal[Filters any, OrderBys any](cursor string) (*OpaqueCursor[Filters, OrderBys], error) {
	var cursorData OpaqueCursor[Filters, OrderBys]

	base64Data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return &cursorData, err
	}

	err = json.Unmarshal(base64Data, &cursorData)
	if err != nil {
		return &cursorData, err
	}

	return &cursorData, nil
}

func UnmarshalWithDefaults[Filters any, OrderBys any](cursor string, defaults OpaqueCursor[Filters, OrderBys]) (*OpaqueCursor[Filters, OrderBys], error) {
	if cursor == "" {
		return &defaults, nil
	}

	var cursorData OpaqueCursor[Filters, OrderBys]

	base64Data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return &cursorData, err
	}

	err = json.Unmarshal(base64Data, &cursorData)
	if err != nil {
		return &defaults, err
	}

	return &cursorData, nil
}

func Marshal[Filters any, OrderBys any](filters Filters, orderBys OrderBys) string {
	cursorData := OpaqueCursor[Filters, OrderBys]{
		Filters:  filters,
		OrderBys: orderBys,
	}

	cursor, err := json.Marshal(cursorData)
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(cursor)
}

type OpaqueCursorUnmarshalError struct {
	Field string
	Err   error
}

func (e *OpaqueCursorUnmarshalError) Error() string {
	return "invalid" + e.Field + "cursor" + ": " + e.Err.Error()
}

type OpaqueCursorPairs[F any, O any] struct {
	Before *OpaqueCursor[F, O]
	After  *OpaqueCursor[F, O]
}

func NewOpaqueCursorPairs[F any, O any](before string, after string) (*OpaqueCursorPairs[F, O], error) {
	var beforeCursor *OpaqueCursor[F, O]
	var err error

	if before != "" {
		beforeCursor, err = Unmarshal[F, O](before)
		if err != nil {
			return nil, &OpaqueCursorUnmarshalError{Field: "before", Err: err}
		}
	}

	var afterCursor *OpaqueCursor[F, O]
	if after != "" {
		afterCursor, err = Unmarshal[F, O](after)
		if err != nil {
			return nil, &OpaqueCursorUnmarshalError{Field: "after", Err: err}
		}
	}

	return &OpaqueCursorPairs[F, O]{
		Before: beforeCursor,
		After:  afterCursor,
	}, nil
}
