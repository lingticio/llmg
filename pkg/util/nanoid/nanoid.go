package nanoid

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/samber/lo"
)

func New() string {
	return lo.Must(gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 8)) //nolint:mnd
}

func NewWithLength(length int) string {
	return lo.Must(gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", length))
}
