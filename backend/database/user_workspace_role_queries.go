package database

import (
	"log"
	"time"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

func GetUserWorkspaceRoles(requester_user_id uint, workspace_id uint) ([]UserWorkspaceRole, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return []UserWorkspaceRole{}, err
	} else if user_role == NoRole {
		return []UserWorkspaceRole{}, custom_errors.ErrAccessDenied
	}

	rows, err := DB.Query(`
		SELECT *
		FROM UserWorkspaceRole
		WHERE workspace_id = $1;`,
		workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return []UserWorkspaceRole{}, custom_errors.ErrDatabaseFailure
	}

	var user_workspace_roles []UserWorkspaceRole

	for rows.Next() {
		var user_workspace_role UserWorkspaceRole

		if err := rows.Scan(
			&user_workspace_role.ID,
			&user_workspace_role.UserID,
			&user_workspace_role.WorkspaceID,
			&user_workspace_role.Role,
			&user_workspace_role.CreatedAt,
			&user_workspace_role.UpdatedAt); err != nil {
			log.Println("Error:", err)
			return []UserWorkspaceRole{}, custom_errors.ErrDatabaseFailure
		} else {
			user_workspace_roles = append(user_workspace_roles, user_workspace_role)
		}
	}

	return user_workspace_roles, nil
}

func AddUserWorkspaceRole(requester_user_id uint, user_id uint, workspace_id uint, role string) (UserWorkspaceRole, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return UserWorkspaceRole{}, err
	} else if user_role != Admin && user_role != Owner {
		return UserWorkspaceRole{}, custom_errors.ErrAccessDenied
	}

	var user_workspace_role UserWorkspaceRole

	err = DB.QueryRow(`
		INSERT INTO
		UserWorkspaceRole(user_id, workspace_id, role, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5)
		RETURNING *;`,
		user_id, workspace_id, role,
		time.Now(), time.Now()).
		Scan(&user_workspace_role.ID,
			&user_workspace_role.UserID,
			&user_workspace_role.WorkspaceID,
			&user_workspace_role.Role,
			&user_workspace_role.CreatedAt,
			&user_workspace_role.UpdatedAt)

	if err != nil {
		log.Println("Error:", err)
		return UserWorkspaceRole{}, custom_errors.ErrDatabaseFailure
	}

	return user_workspace_role, nil
}

func UpdateUserWorkspaceRole(requester_user_id uint, user_id uint, workspace_id uint, role string) error {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	target_user_role, err := getUserWorkspaceRole(user_id, workspace_id)
	if err != nil {
		return err
	} else if target_user_role == NoRole {
		return custom_errors.ErrInvalidArguments
	} else if target_user_role == Owner {
		return custom_errors.ErrAccessDenied
	} else if user_role == Admin && target_user_role != StandardUser {
		return custom_errors.ErrAccessDenied
	}

	if role == Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		UPDATE UserWorkspaceRole
		SET role = $1, updated_at = $2
		WHERE user_id = $3 AND workspace_id = $4;`,
		role, time.Now(), user_id, workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func DeleteUserWorkspaceRole(requester_user_id uint, user_id uint, workspace_id uint) error {
	if requester_user_id != user_id {
		user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
		if err != nil {
			return err
		} else if user_role != Admin && user_role != Owner {
			return custom_errors.ErrAccessDenied
		}

		target_user_role, err := getUserWorkspaceRole(user_id, workspace_id)
		if err != nil {
			return err
		} else if target_user_role == NoRole {
			return custom_errors.ErrInvalidArguments
		} else if target_user_role == Owner {
			return custom_errors.ErrInvalidArguments
		} else if user_role == Admin && target_user_role != StandardUser {
			return custom_errors.ErrAccessDenied
		}
	}

	_, err := DB.Exec(`
		DELETE FROM UserWorkspaceRole
		WHERE user_id = $1 AND workspace_id = $2;`,
		user_id, workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}
