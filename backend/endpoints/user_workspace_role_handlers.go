package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
	"github.com/skye-tan/trello/backend/utils/custom_messages"
	"github.com/skye-tan/trello/backend/websocket_utils"
)

// GET "/workspaces/:workspace_id/users"
func getUserWorkspaceRoles(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	user_workespace_roles, err := database.GetUserWorkspaceRoles(requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, user_workespace_roles)
}

// POST "/workspaces/:workspace_id/users"
func addUserWorkspaceRole(c echo.Context) error {
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

	tmp, ok := content["user_id"].(string)
	user_id, err := strconv.ParseUint(tmp, 10, 32)
	if !ok || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	role, ok := content["role"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}
	if role != database.Admin && role != database.Owner && role != database.StandardUser {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidRole)
	}

	user_workespace_role, err := database.AddUserWorkspaceRole(requester_user_id, uint(user_id), workspace_id, role)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	workspace, err := database.GetWorkspace(requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	members, err := database.GetWorkspaceMembers(workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: members,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.MemeberGroup,
			Type:    websocket_utils.UpdateType,
			Message: "Member has been added.",
		},
	}

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: []uint{uint(user_id)},
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.WorkspaceGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("You have been added to workspace '%s'.", workspace.Name),
		},
	}

	return c.JSON(http.StatusCreated, user_workespace_role)
}

// PUT "/workspaces/:workspace_id/users/:user_id"
func updateUserWorkspaceRole(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	user_id, ok := extractQueryParameter(c, "user_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidUserId)
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

	role, ok := content["role"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}
	if role != database.Admin && role != database.Owner && role != database.StandardUser {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidRole)
	}

	err = database.UpdateUserWorkspaceRole(requester_user_id, user_id, workspace_id, role)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	workspace, err := database.GetWorkspace(requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	members, err := database.GetWorkspaceMembers(workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: members,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.MemeberGroup,
			Type:    websocket_utils.UpdateType,
			Message: "Member has been updated.",
		},
	}

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: []uint{user_id},
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.WorkspaceGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("Your role in workspace '%s' has been updated.", workspace.Name),
		},
	}

	return c.NoContent(http.StatusCreated)
}

// DELETE "/workspaces/:workspace_id/users/:user_id"
func deleteUserWorkspaceRole(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	user_id, ok := extractQueryParameter(c, "user_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidUserId)
	}

	err := database.DeleteUserWorkspaceRole(requester_user_id, user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	members, err := database.GetWorkspaceMembers(workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: members,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.MemeberGroup,
			Type:    websocket_utils.UpdateType,
			Message: "Member has been removed.",
		},
	}

	workspace, err := database.GetWorkspace(requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: []uint{user_id},
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.WorkspaceGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("You have been removed from workspace '%s'.", workspace.Name),
		},
	}

	return c.NoContent(http.StatusOK)
}

// DELETE "/workspaces/:workspace_id/users/leave"
func leaveUserWorkspaceRole(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	err := database.DeleteUserWorkspaceRole(requester_user_id, requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	members, err := database.GetWorkspaceMembers(workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: members,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.MemeberGroup,
			Type:    websocket_utils.UpdateType,
			Message: "Member has been removed.",
		},
	}

	return c.NoContent(http.StatusOK)
}
