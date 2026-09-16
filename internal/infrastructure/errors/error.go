package errors

import (
	"github.com/anderson-reinaldo/go-explicAI/internal/application"
	echo "github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func Handle(c echo.Context, err error) error {
	switch errors.Cause(err) {
	case application.MissingFile, application.InvalidFile:
		return echo.ErrBadRequest
	case application.FailedReadFile:
		return echo.ErrUnprocessableEntity
	default:
		return echo.ErrInternalServerError
	}
}
