package handlers

import (
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/skye-tan/trello/backend/middlewares/authentication"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
)

func customLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Path() == "/metrics" {
			return next(c)
		}
		return middleware.Logger()(next)(c)
	}
}

func Start(listen_address string) {
	e := echo.New()
	e.Use(customLogger)

	// Metrics Endpoint
	e.GET("/metrics", echoprometheus.NewHandlerWithConfig(
		echoprometheus.HandlerConfig{
			Gatherer: monitoring.Registry,
		}),
	)

	api := e.Group("/api")

	// Statistics Collector Middleware
	api.Use(monitoring.StatisticsCollectorMiddleware())

	// Miscellaneous Endpoints
	api.POST("/signup", signup)
	api.POST("/login", login)
	api.GET("/ws/:token", websocketHandler)
	api.GET("/token/validate", validateToken, authentication.AccessJWTMiddleware)
	api.GET("/token/refresh", refreshToken, authentication.RefreshJWTMiddleware)

	// Workspace Endpoints
	api.GET("/workspaces", getWorkspaces, authentication.AccessJWTMiddleware)
	api.POST("/workspaces", createWorkspace, authentication.AccessJWTMiddleware)
	api.GET("/workspaces/:workspace_id", getWorksapce, authentication.AccessJWTMiddleware)
	api.PUT("/workspaces/:workspace_id", updateWorkspace, authentication.AccessJWTMiddleware)
	api.DELETE("/workspaces/:workspace_id", deleteWorkspace, authentication.AccessJWTMiddleware)

	// Task Endpoints
	api.GET("/self/tasks", getAssignedTasks, authentication.AccessJWTMiddleware)
	api.GET("/workspaces/:workspace_id/tasks", getTasks, authentication.AccessJWTMiddleware)
	api.POST("/workspaces/:workspace_id/tasks", createTask, authentication.AccessJWTMiddleware)
	api.GET("/workspaces/:workspace_id/tasks/:task_id", getTask, authentication.AccessJWTMiddleware)
	api.PUT("/workspaces/:workspace_id/tasks/:task_id", updateTask, authentication.AccessJWTMiddleware)
	api.PUT("/workspaces/:workspace_id/tasks/:task_id/status", updateTaskStatus, authentication.AccessJWTMiddleware)
	api.DELETE("/workspaces/:workspace_id/tasks/:task_id", deleteTask, authentication.AccessJWTMiddleware)

	// Subtask Endpoints
	api.GET("/tasks/:task_id/subtasks", getSubtasks, authentication.AccessJWTMiddleware)
	api.POST("/tasks/:task_id/subtasks", createSubtask, authentication.AccessJWTMiddleware)
	api.GET("/tasks/:task_id/subtasks/:subtask_id", getSubtask, authentication.AccessJWTMiddleware)
	api.PUT("/tasks/:task_id/subtasks/:subtask_id", updateSubtask, authentication.AccessJWTMiddleware)
	api.PUT("/tasks/:task_id/subtasks/:subtask_id/status", updateSubtaskStatus, authentication.AccessJWTMiddleware)
	api.PUT("/tasks/:task_id/subtasks/:subtask_id/title", updateSubtaskTitle, authentication.AccessJWTMiddleware)
	api.PUT("/tasks/:task_id/subtasks/:subtask_id/assigneeid", updateSubtaskAssigneeID, authentication.AccessJWTMiddleware)
	api.DELETE("/tasks/:task_id/subtasks/:subtask_id", deleteSubtask, authentication.AccessJWTMiddleware)

	// User-Profile Endpoints
	api.GET("/users/self/profile", getSelfProfile, authentication.AccessJWTMiddleware)
	api.GET("/users/:user_id/profile/id", getUserProfileByID, authentication.AccessJWTMiddleware)
	api.GET("/users/:username/profile/username", getUserProfileByUsername, authentication.AccessJWTMiddleware)
	api.PUT("/users/self/profile/username", updateSelfProfileUsername, authentication.AccessJWTMiddleware)
	api.PUT("/users/self/profile/password", updateSelfProfilePassword, authentication.AccessJWTMiddleware)
	api.DELETE("/users/self/profile", deleteSelfProfile, authentication.AccessJWTMiddleware)

	// User-Workspace-Role Endpoints
	api.GET("/workspaces/:workspace_id/users", getUserWorkspaceRoles, authentication.AccessJWTMiddleware)
	api.POST("/workspaces/:workspace_id/users", addUserWorkspaceRole, authentication.AccessJWTMiddleware)
	api.PUT("/workspaces/:workspace_id/users/:user_id", updateUserWorkspaceRole, authentication.AccessJWTMiddleware)
	api.DELETE("/workspaces/:workspace_id/users/:user_id", deleteUserWorkspaceRole, authentication.AccessJWTMiddleware)
	api.DELETE("/workspaces/:workspace_id/users/leave", leaveUserWorkspaceRole, authentication.AccessJWTMiddleware)

	// Comment Endpoints
	api.GET("/workspaces/:workspace_id/tasks/:task_id/comments", getComments, authentication.AccessJWTMiddleware)
	api.POST("/workspaces/:workspace_id/tasks/:task_id/comments", addComment, authentication.AccessJWTMiddleware)

	// Watch Endpoints
	api.GET("/workspaces/:workspace_id/tasks/:task_id/watch", getWatch, authentication.AccessJWTMiddleware)
	api.POST("/workspaces/:workspace_id/tasks/:task_id/watch", addWatch, authentication.AccessJWTMiddleware)
	api.DELETE("/workspaces/:workspace_id/tasks/:task_id/watch", deleteWatch, authentication.AccessJWTMiddleware)

	// Upload Picture Endpoints
	api.POST("/upload/picture/:task_id", uploadFile, authentication.AccessJWTMiddleware)
	api.GET("/retrieve/picture/:task_id", retrieveFile)

	e.Logger.Fatal(e.Start(listen_address))
}
