package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

func getUserIDFromContext(c echo.Context) uint {
	user_id, ok := c.Get("user_id").(uint)
	if !ok {
		panic("Failed to get user_id from context.")
	}
	return user_id
}

func extractQueryParameter(c echo.Context, parameter string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(parameter), 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(value), true
}

func generateProperResponse(err error) *echo.HTTPError {
	message := err.Error()
	if err == custom_errors.ErrTokenFailure {
		return echo.NewHTTPError(http.StatusInternalServerError, message)
	} else if err == custom_errors.ErrAccessDenied {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, message)
	}
}
