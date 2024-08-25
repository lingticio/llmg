package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lingticio/gateway/pkg/apierrors"
)

func NotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, apierrors.NewErrNotFound().AsResponse())
}
