package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
	"github.com/skye-tan/trello/backend/utils/custom_messages"
	"github.com/skye-tan/trello/backend/websocket_utils"
)

// GET "/self/tasks"
func getAssignedTasks(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	tasks, err := database.GetAssignedTasks(requester_user_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, tasks)
}

// GET "/workspaces/:workspace_id/tasks"
func getTasks(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	tasks, err := database.GetAllTasksInWorkspace(requester_user_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, tasks)
}

// POST "/workspaces/:workspace_id/tasks"
func createTask(c echo.Context) error {
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

	title, ok := content["title"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	description, ok := content["description"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp, ok := content["estimated_time"].(string)
	estimated_time, err := strconv.Atoi(tmp)
	if !ok || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp, ok = content["actual_time"].(string)
	actual_time, err := strconv.Atoi(tmp)

	if !ok || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp, ok = content["due_date"].(string)
	due_date, err := time.Parse("2006-01-02", tmp)
	if !ok || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp, ok = content["priority"].(string)
	priority, err := strconv.Atoi(tmp)
	if !ok || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp, ok = content["assignee_id"].(string)
	assignee_id, err := strconv.Atoi(tmp)
	if !ok || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	image_url, ok := content["image_url"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	task, err := database.CreateTaskInWorkspace(requester_user_id, workspace_id, title, description, estimated_time, actual_time, due_date, priority, uint(assignee_id), image_url)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: []uint{uint(assignee_id)},
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("Task '%s' has been assigned to you.", task.Title),
		},
	}

	associated_users, err := database.GetAssociatedUsersWithTask(task.ID)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Task '%s' has been created.", task.Title),
		},
	}

	return c.JSON(http.StatusCreated, task)
}

// GET "/workspaces/:workspace_id/tasks/:task_id"
func getTask(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	task, err := database.GetDetailsOfTask(requester_user_id, workspace_id, task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, task)
}

// PUT "/workspaces/:workspace_id/tasks/:task_id/status
func updateTaskStatus(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

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

	status, ok := content["status"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	} else if status != database.Planned && status != database.Completed && status != database.InProgress {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidStatus)
	}

	task, err := database.GetDetailsOfTask(requester_user_id, workspace_id, task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.UpdateStatusOfTask(requester_user_id, workspace_id, task_id, status)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	watchers, err := database.GetWatchers(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: watchers,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("Task '%s' has been updated.", task.Title),
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
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Task '%s' has been updated.", task.Title),
		},
	}

	return c.NoContent(http.StatusCreated)
}

// PUT "/workspaces/:workspace_id/tasks/:task_id"
func updateTask(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

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

	description, ok := content["description"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp2, ok := content["actual_time"].(float64)
	actual_time := int(tmp2)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp, ok := content["due_date"].(string)
	due_date, err := time.Parse("2006-01-02", strings.Split(tmp, "T")[0])
	if !ok || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp2, ok = content["priority"].(float64)
	priority := int(tmp2)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	tmp3, ok := content["assignee_id"].(string)
	tmp4, ok2 := content["assignee_id"].(float64)
	tmp5, _ := strconv.Atoi(tmp3)
	assignee_id := uint(0)
	if ok {
		assignee_id = uint(tmp5)
	} else {
		assignee_id = uint(tmp4)
	}
	if !ok && !ok2 {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	image_url, ok := content["image_url"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	task, err := database.GetDetailsOfTask(requester_user_id, workspace_id, task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.UpdateDetailsOfTask(requester_user_id, workspace_id, task_id, title, description, actual_time, due_date, priority, assignee_id, image_url)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	if assignee_id != task.AssigneeID {
		websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
			TargetUserIDs: []uint{uint(assignee_id)},
			Body: &websocket_utils.WebsocketBody{
				Group:   websocket_utils.TaskGroup,
				Type:    websocket_utils.WatchType,
				Message: fmt.Sprintf("Task '%s' has been assigned to you.", task.Title),
			},
		}
	}

	watchers, err := database.GetWatchers(task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: watchers,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("Task '%s' has been updated.", task.Title),
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
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Task '%s' has been updated.", task.Title),
		},
	}

	return c.NoContent(http.StatusCreated)
}

// DELETE "/workspaces/:workspace_id/tasks/:task_id"
func deleteTask(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	task, err := database.GetDetailsOfTask(requester_user_id, workspace_id, task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	err = database.DeleteTask(requester_user_id, workspace_id, task_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	watchers, err := database.GetWatchers(task_id)
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
		TargetUserIDs: watchers,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.WatchType,
			Message: fmt.Sprintf("Task '%s' has been deleted.", task.Title),
		},
	}

	websocket_utils.Hub.Broadcast <- &websocket_utils.WebsocketBroadcast{
		TargetUserIDs: associated_users,
		Body: &websocket_utils.WebsocketBody{
			Group:   websocket_utils.TaskGroup,
			Type:    websocket_utils.UpdateType,
			Message: fmt.Sprintf("Task '%s' has been deleted.", task.Title),
		},
	}

	return c.NoContent(http.StatusCreated)
}
