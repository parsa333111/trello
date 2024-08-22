package authentication

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/utils/custom_messages"
	token_utils "github.com/skye-tan/trello/backend/utils/token"
)

func AccessJWTMiddleware(handler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth_header := c.Request().Header.Get("Authorization")
		if auth_header == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.MissingToken)
		}

		split_token := strings.Split(auth_header, "Bearer ")
		if len(split_token) != 2 {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidTokenFormat)
		}

		token := split_token[1]
		claims, err := token_utils.ParseToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidToken)
		} else if claims.Type != token_utils.Access {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidTokenType)
		}

		_, err = database.GetUserByID(claims.UserID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidToken)
		}

		c.Set("user_id", claims.UserID)
		return handler(c)
	}
}

func RefreshJWTMiddleware(handler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("RefreshToken")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.MissingToken)
		}

		claims, err := token_utils.ParseToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidToken)
		} else if claims.Type != token_utils.Refresh {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidTokenType)
		}

		_, err = database.GetUserByID(claims.UserID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidToken)
		}

		c.Set("user_id", claims.UserID)
		return handler(c)
	}
}
