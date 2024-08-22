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

// GET "/tasks/:task_id/subtasks"
func getSubtasks(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	subtasks, err := database.GetAllSubtasksInTask(requester_user_id, task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, subtasks)
}

// POST "/tasks/:task_id/subtasks"
func createSubtask(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
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

	title, ok := content["title"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp, ok := content["assignee_id"].(string)
	assignee_id, _ := strconv.Atoi(tmp)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	subtask, err := database.CreateSubtaskInTask(requester_user_id, task_id, title, uint(assignee_id))
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: []uint{uint(assignee_id)},
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.SubtaskGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("Subtask '%s' has been assigned to you.", subtask.Title),
		},
	}

	associated_users, err := database.GetAssociatedUsersWithTask(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.SubtaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Subtask '%s' has been created.", subtask.Title),
		},
	}

	return c.JSON(http.StatusCreated, subtask)
}

// GET "/tasks/:task_id/subtasks/:subtask_id"
func getSubtask(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	subtask_id, ok := extractQueryParameter(c, "subtask_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidSubtaskId)
	}

	subtask, err := database.GetDetailsOfSubtask(requester_user_id, task_id, subtask_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, subtask)
}

// PUT "/tasks/:task_id/subtasks/:subtask_id"
func updateSubtask(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	subtask_id, ok := extractQueryParameter(c, "subtask_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidSubtaskId)
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

	title, ok := content["title"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	is_completed, ok := content["is_completed"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	} else if is_completed != database.Yes && is_completed != database.No {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidIsCompleted)
	}

	assignee_id, ok := content["assignee_id"].(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	subtask, err := database.GetDetailsOfSubtask(requester_user_id, task_id, subtask_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.UpdateDetailsOfSubtask(requester_user_id, task_id, subtask_id, title, is_completed, assignee_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	if assignee_id != subtask.AssigneeID {
		websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
			TargetUserIDs: []uint{uint(assignee_id)},
			Body: &websocket_utils.WebsocketBody{
				Group:   websocket_utils.SubtaskGroup,
				Type:    websocket_utils.WatchType,
				Message: fmt.Sprintf("Subtask '%s' has been assigned to you.", subtask.Title),
			},
		}
	}

	associated_users, err := database.GetAssociatedUsersWithTask(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.SubtaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Subtask '%s' has been created.", subtask.Title),
		},
	}

	return c.NoContent(http.StatusCreated)
}

// PUT "/tasks/:task_id/subtasks/:subtask_id/assigneeid"
func updateSubtaskAssigneeID(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	subtask_id, ok := extractQueryParameter(c, "subtask_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidSubtaskId)
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

	tmp, ok := content["assignee_id"].(string)
	tmp2, _ := strconv.Atoi(tmp)
	assignee_id := uint(tmp2)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	subtask, err := database.GetDetailsOfSubtask(requester_user_id, task_id, subtask_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.UpdateDetailsOfSubtaskAssigneeID(requester_user_id, task_id, subtask_id, assignee_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	if assignee_id != subtask.AssigneeID {
		websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
			TargetUserIDs: []uint{uint(assignee_id)},
			Body: &websocket_utils.WebsocketBody{
				Group:   websocket_utils.SubtaskGroup,
				Type:    websocket_utils.WatchType,
				Message: fmt.Sprintf("Subtask '%s' has been assigned to you.", subtask.Title),
			},
		}
	}

	associated_users, err := database.GetAssociatedUsersWithTask(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.SubtaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Subtask '%s' has been updated.", subtask.Title),
		},
	}

	return c.NoContent(http.StatusCreated)
}

// PUT "/tasks/:task_id/subtasks/:subtask_id/status"
func updateSubtaskStatus(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	subtask_id, ok := extractQueryParameter(c, "subtask_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidSubtaskId)
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

	is_completed, ok := content["is_completed"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	} else if is_completed != database.Yes && is_completed != database.No {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidIsCompleted)
	}

	subtask, err := database.GetDetailsOfSubtask(requester_user_id, task_id, subtask_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.UpdateDetailsOfSubtaskStatus(requester_user_id, task_id, subtask_id, is_completed)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	associated_users, err := database.GetAssociatedUsersWithTask(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.SubtaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Subtask '%s' has been updated.", subtask.Title),
		},
	}

	return c.NoContent(http.StatusCreated)
}

// PUT "/tasks/:task_id/subtasks/:subtask_id/title"
func updateSubtaskTitle(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	subtask_id, ok := extractQueryParameter(c, "subtask_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidSubtaskId)
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

	title, ok := content["title"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	subtask, err := database.GetDetailsOfSubtask(requester_user_id, task_id, subtask_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.UpdateDetailsOfSubtaskTitle(requester_user_id, task_id, subtask_id, title)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	associated_users, err := database.GetAssociatedUsersWithTask(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.SubtaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Subtask '%s' has been updated.", subtask.Title),
		},
	}

	return c.NoContent(http.StatusCreated)
}

// DELETE "/tasks/:task_id/subtasks/:subtask_id"
func deleteSubtask(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	subtask_id, ok := extractQueryParameter(c, "subtask_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidSubtaskId)
	}

	subtask, err := database.GetDetailsOfSubtask(requester_user_id, task_id, subtask_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	associated_users, err := database.GetAssociatedUsersWithTask(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.DeleteSubtask(requester_user_id, task_id, subtask_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.SubtaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Subtask '%s' has been deleted.", subtask.Title),
		},
	}

	return c.NoContent(http.StatusOK)
}
