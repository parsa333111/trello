package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
	"github.com/skye-tan/trello/backend/utils/custom_messages"
	"github.com/skye-tan/trello/backend/websocket_utils"
)

// GET "/workspaces/:workspace_id/tasks/:task_id/comments"
func getComments(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	comments, err := database.GetComments(requester_user_id, task_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.JSON(http.StatusOK, comments)
}

// POST "/workspaces/:workspace_id/tasks/:task_id/comments"
func addComment(c echo.Context) error {
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

	text, ok := content["text"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.MissingData)
	}

	comment, err := database.AddComment(requester_user_id, task_id, workspace_id, text)
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
			Group:   websocket_utils.CommentGroup,
			Type:    websocket_utils.UpdateType,
			Message: text,
		},
	}

	return c.JSON(http.StatusCreated, comment)
}

// GET "/workspaces/:workspace_id/tasks/:task_id/watch"
func getWatch(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	response, err := database.GetWatch(requester_user_id, task_id, workspace_id)

	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	var status database.WatchStatus
	status.Status = response

	return c.JSON(http.StatusOK, status)
}

// POST "/workspaces/:workspace_id/tasks/:task_id/watch"
func addWatch(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	err := database.AddWatch(requester_user_id, task_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusCreated)
}

// DELETE "/workspaces/:workspace_id/tasks/:task_id/watch"
func deleteWatch(c echo.Context) error {
	requester_user_id := getUserIDFromContext(c)

	workspace_id, ok := extractQueryParameter(c, "workspace_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidWorkspaceId)
	}

	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	err := database.DeleteWatch(requester_user_id, task_id, workspace_id)
	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
		return generateProperResponse(err)
	}
	monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

	return c.NoContent(http.StatusOK)
}

// POST "/upload/picture/:task_id"
func uploadFile(c echo.Context) error {
	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	task_id_string := strconv.Itoa(int(task_id))

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, custom_messages.PictureCreateFailure)
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.PictureOpenFailure)
	}
	defer src.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, custom_messages.PictureSaveFailure)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	err = database.AddPictureToRedis(task_id_string, encoded)

	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.PictureSaveFailure)
	}
	return echo.NewHTTPError(http.StatusNoContent)
}

// GET "/retrieve/picture/task_id"
func retrieveFile(c echo.Context) error {
	task_id, ok := extractQueryParameter(c, "task_id")
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, custom_messages.InvalidTaskId)
	}

	task_id_string := strconv.Itoa(int(task_id))

	decoded, err := database.RetrieveFile(task_id_string)

	if err != nil {
		return c.JSON(http.StatusNoContent, "Nothing to retrieve")
	}

	return c.Blob(http.StatusOK, "image/jpeg", decoded)
}
