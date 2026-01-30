package apperr

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GlobalErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var ve *ValidationError
		if errors.As(err, &ve) {
			_ = c.JSON(http.StatusBadRequest, map[string]string{"error": ve.Message, "title": "validation error"})
			return
		}

		var he *echo.HTTPError
		if errors.As(err, &he) {
			msg := fmt.Sprintf("%v", he.Message)
			_ = c.JSON(he.Code, map[string]string{"error": msg})
			return
		}

		slog.Error("Unhandled error", "error", err)
		_ = c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
