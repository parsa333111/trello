package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
	"github.com/skye-tan/trello/backend/utils/custom_messages"
	hashing_utils "github.com/skye-tan/trello/backend/utils/hashing"
	regex_utils "github.com/skye-tan/trello/backend/utils/regex"
	"github.com/skye-tan/trello/backend/websocket_utils"
)

// GET "/users/self/profile"
func getSelfProfile(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	user, err := database.GetUserByID(requester_user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, map[string]string{
		"id":         fmt.Sprint(user.ID),
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format("2000-January-01"),
		"updated_at": user.UpdatedAt.Format("2000-January-01"),
	})
}

// GET "/users/:user_id/profile/id"
func getUserProfileByID(c echo.Context) error {
	user_id, ok := extractQueryParameter(c, "user_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidUserId)
	}

	user, err := database.GetUserByID(user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	responder := make(chan string)
	websocket_utils.Hub.GetStatus <- &websocket_utils.WebsocketGetStatus{
		TargetUserID: user_id,
		Responder:    responder,
	}
	user_status := <-responder

	return c.JSON(http.StatusOK, map[string]string{
		"id":         fmt.Sprint(user.ID),
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format("2000-January-01"),
		"status":     user_status,
	})
}

// GET "/users/:username/profile/username"
func getUserProfileByUsername(c echo.Context) error {
	username := c.Param("username")

	user, err := database.GetUserByUsername(username)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	responder := make(chan string)
	websocket_utils.Hub.GetStatus <- &websocket_utils.WebsocketGetStatus{
		TargetUserID: user.ID,
		Responder:    responder,
	}
	user_status := <-responder

	return c.JSON(http.StatusOK, map[string]string{
		"id":         fmt.Sprint(user.ID),
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format("2000-January-01"),
		"status":     user_status,
	})
}

// PUT "/users/self/profile/username"
func updateSelfProfileUsername(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	content_type := c.Request().Header.Get(echo.HeaderContentType)
	if content_type != "application/json" {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidContentType)
	}

	content := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&content)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidBodyFormat)
	}

	user, err := database.GetUserByID(requester_user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	username, ok := content["username"].(string)
	if !ok {
		username = user.Username
	} else if !regex_utils.ValidateUsername(username) {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidUsernameFormat)
	}

	err = database.UpdateUserUsername(requester_user_id, username)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusCreated)
}

// PUT "/users/self/profile/password"
func updateSelfProfilePassword(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	content_type := c.Request().Header.Get(echo.HeaderContentType)
	if content_type != "application/json" {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidContentType)
	}

	content := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&content)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidBodyFormat)
	}

	user, err := database.GetUserByID(requester_user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	var password_hash []byte
	password, ok := content["password"].(string)
	if !ok {
		password_hash = user.PasswordHash
	} else if !regex_utils.ValidatePassword(password) {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidPasswordFormat)
	} else {
		password_hash = hashing_utils.HashUsingSha256(password)
	}

	err = database.UpdateUserPassword(requester_user_id, password_hash)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusCreated)
}

// DELETE "/users/self/profile"
func deleteSelfProfile(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	err := database.DeleteUser(requester_user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusOK)
}
