package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
	"github.com/skye-tan/trello/backend/utils/custom_messages"
	hashing_utils "github.com/skye-tan/trello/backend/utils/hashing"
	regex_utils "github.com/skye-tan/trello/backend/utils/regex"
	token_utils "github.com/skye-tan/trello/backend/utils/token"
	"github.com/skye-tan/trello/backend/websocket_utils"
)

// POST "/signup"
func signup(c echo.Context) error {
	content_type := c.Request().Header.Get(echo.HeaderContentType)
	if content_type != "application/json" {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidContentType)
	}

	content := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&content)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidBodyFormat)
	}

	username, ok := content["username"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	} else if !regex_utils.ValidateUsername(username) {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidUsernameFormat)
	}

	email, ok := content["email"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	} else if !regex_utils.ValidateEmail(email) {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidEmailFormat)
	}

	password, ok := content["password"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	} else if !regex_utils.ValidatePassword(password) {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidPasswordFormat)
	}

	err = database.CreateUser(username, email, hashing_utils.HashUsingSha256(password))
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusCreated)
}

// POST "/login"
func login(c echo.Context) error {
	content_type := c.Request().Header.Get(echo.HeaderContentType)
	if content_type != "application/json" {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidContentType)
	}

	content := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&content)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidBodyFormat)
	}

	username, ok := content["username"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	password, ok := content["password"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	user, err := database.GetUserByUsername(username)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidCredentials)
	}

	if !bytes.Equal(user.PasswordHash, hashing_utils.HashUsingSha256(password)) {
		return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidCredentials)
	}

	refresh_token, access_token, err := token_utils.GenerateTokens(user.ID)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, map[string]string{
		"AccessToken":  access_token,
		"RefreshToken": refresh_token,
	})
}

// GET "/ws/:token"
func websocketHandler(c echo.Context) error {
	token := c.Param("token")

	claims, err := token_utils.ParseToken(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidToken)
	} else if claims.Type != token_utils.Access {
		return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidTokenType)
	}

	user_id := claims.UserID

	_, err = database.GetUserByID(user_id)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, custom_messages.InvalidToken)
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	websocket_utils.Hub.Register <- &websocket_utils.WebsocketConn{
		UserID: claims.UserID,
		Conn:   conn,
	}

	associated_users, err := database.GetAssociatedUsersWithUser(user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.MemeberGroup,
			Type:    websocket_utils.UpdateType,
			Message: "A user is online now.",
		},
	}

	go websocket_utils.HandleClientWebsocket(conn, claims.UserID)

	return c.NoContent(http.StatusOK)
}

// GET "/token/validate"
func validateToken(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

// GET "/token/refresh"
func refreshToken(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	access_token, err := token_utils.GenerateToken(requester_user_id, token_utils.Access)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, map[string]string{
		"AccessToken": access_token,
	})
}
