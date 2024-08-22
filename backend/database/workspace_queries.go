package database

import (
	"log"
	"time"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

func GetWorkspaces(requester_user_id uint) ([]Workspace, error) {
	rows, err := DB.Query(`
		SELECT A.*
		FROM Workspace A JOIN UserWorkspaceRole B
		ON A.id = B.workspace_id
		WHERE B.user_id = $1;`,
		requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return []Workspace{}, custom_errors.ErrDatabaseFailure
	}

	var workspaces []Workspace

	for rows.Next() {
		var workspace Workspace

		if err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.CreatedAt,
			&workspace.UpdatedAt); err != nil {
			log.Println("Error:", err)
			return []Workspace{}, custom_errors.ErrDatabaseFailure
		} else {
			workspaces = append(workspaces, workspace)
		}
	}

	return workspaces, nil
}

func CreateWorkSpace(requester_user_id uint, name string, description string) (Workspace, error) {
	if ok := checkDuplicateWorkspaceName(name); !ok {
		return Workspace{}, custom_errors.ErrDuplicateWorspaceName
	}

	var workspace Workspace

	err := DB.QueryRow(`
		INSERT INTO
		Workspace(name, description, created_at, updated_at) 
		VALUES($1, $2, $3, $4)
		RETURNING *;`,
		name, description, time.Now(), time.Now()).
		Scan(&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.CreatedAt,
			&workspace.UpdatedAt)

	if err != nil {
		log.Println("Error:", err)
		return Workspace{}, custom_errors.ErrDatabaseFailure
	}

	_, err = DB.Exec(`
		INSERT INTO
		UserWorkspaceRole(user_id, workspace_id, role, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5);`,
		requester_user_id, workspace.ID, Owner,
		time.Now(), time.Now())

	if err != nil {
		log.Println("Error:", err)
		return Workspace{}, custom_errors.ErrDatabaseFailure
	}

	return workspace, nil
}

func GetWorkspace(requester_user_id uint, workspace_id uint) (Workspace, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return Workspace{}, err
	} else if user_role == NoRole {
		return Workspace{}, custom_errors.ErrAccessDenied
	}

	var workspace Workspace

	err = DB.QueryRow(`
		SELECT B.*
		FROM UserWorkspaceRole A JOIN workspace B
		ON A.workspace_id = B.id
		WHERE A.user_id = $1 AND B.id = $2;`,
		requester_user_id, workspace_id).
		Scan(&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.CreatedAt,
			&workspace.UpdatedAt)

	if err != nil {
		log.Println("Error:", err)
		return Workspace{}, custom_errors.ErrDatabaseFailure
	}

	return workspace, nil
}

func UpdateWorkspace(requester_user_id uint, workspace_id uint, name string, description string) error {
	if ok := checkDuplicateWorkspaceName(name); !ok {
		return custom_errors.ErrDuplicateWorspaceName
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		UPDATE Workspace 
		SET name = $1, description = $2, update_at = $3,
		WHERE id = $4;`,
		name, description, time.Now(), workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func DeleteWorkspace(requester_user_id uint, workspace_id uint) error {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		DELETE FROM Workspace 
		WHERE id = $1;`,
		workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}
