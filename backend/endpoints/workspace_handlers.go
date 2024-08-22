package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
	"github.com/skye-tan/trello/backend/utils/custom_messages"
)

// GET "/workspaces"
func getWorkspaces(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	worksapces, err := database.GetWorkspaces(requester_user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, worksapces)
}

// POST "/workspaces"
func createWorkspace(c echo.Context) error {
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

	name, ok := content["name"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	description, ok := content["description"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	worksapce, err := database.CreateWorkSpace(requester_user_id, name, description)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusCreated, worksapce)
}

// GET "/workspaces/:workspace_id"
func getWorksapce(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	worksapce, err := database.GetWorkspace(requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, worksapce)
}

// PUT "/workspaces/:workspace_id"
func updateWorkspace(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	content_type := c.Request().Header.Get(echo.HeaderContentType)
	if content_type != "application/json" {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidContentType)
	}

	content := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&content)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidBodyFormat)
	}

	name, ok := content["name"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	description, ok := content["description"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	err = database.UpdateWorkspace(requester_user_id, workspace_id, name, description)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusCreated)
}

// DELETE "/workspaces/:workspace_id"
func deleteWorkspace(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	err := database.DeleteWorkspace(requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusOK)
}
