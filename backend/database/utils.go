package database

import (
	"log"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

const (
	Owner        = "Owner"
	Admin        = "Admin"
	StandardUser = "StandardUser"
	NoRole       = ""
)

const (
	Planned    = "Planned"
	InProgress = "InProgress"
	Completed  = "Completed"
)

const (
	Yes = "Yes"
	No  = "No"
)

func getUserWorkspaceRole(user_id uint, workspace_id uint) (string, error) {
	var role string

	err := DB.QueryRow(`
		SELECT role
		FROM UserWorkspaceRole
		WHERE user_id = $1 AND workspace_id = $2;`,
		user_id, workspace_id).
		Scan(&role)

	if err != nil {
		log.Println("Error:", err)
		return "", custom_errors.ErrDatabaseFailure
	}

	return role, nil
}

func getTaskWorkspaceID(task_id uint) (uint, error) {
	var workspace_id uint

	err := DB.QueryRow(`
		SELECT workspace_id
		FROM Task 
		WHERE id = $1;`,
		task_id).Scan(&workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return 0, custom_errors.ErrDatabaseFailure
	}

	return workspace_id, nil
}

func checkDuplicateUsername(username string) bool {
	rows, err := DB.Query(`
		SELECT * FROM Users
		WHERE username = $1;`,
		username)

	if rows.Next() || err != nil {
		return false
	}

	return true
}

func checkDuplicateEmail(email string) bool {
	rows, err := DB.Query(`
		SELECT * FROM Users
		WHERE email = $1;`,
		email)

	if rows.Next() || err != nil {
		return false
	}

	return true
}

func checkDuplicateWorkspaceName(name string) bool {
	rows, err := DB.Query(`
		SELECT * FROM Workspace
		WHERE name = $1;`,
		name)

	if rows.Next() || err != nil {
		return false
	}

	return true
}

func checkDuplicateTaskTitle(workspace_id uint, title string) bool {
	rows, err := DB.Query(`
		SELECT * FROM Task
		WHERE workspace_id = $1 AND title = $2`,
		workspace_id, title)

	if rows.Next() || err != nil {
		return false
	}

	return true
}

func checkDuplicateTaskTitleWithoutSelf(workspace_id uint, title string, id uint) bool {
	rows, err := DB.Query(`
		SELECT * FROM Task
		WHERE workspace_id = $1 AND title = $2 AND id != $3`,
		workspace_id, title, id)

	if rows.Next() || err != nil {
		return false
	}

	return true
}

func checkDuplicateSubtaskTitle(task_id uint, title string) bool {
	rows, err := DB.Query(`
		SELECT * FROM Subtask
		WHERE task_id = $1 AND title = $2`,
		task_id, title)

	if rows.Next() || err != nil {
		return false
	}

	return true
}
