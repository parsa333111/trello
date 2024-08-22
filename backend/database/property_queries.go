package database

import (
	"log"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

func GetComments(requester_user_id uint, task_id uint, workspace_id uint) ([]Comment, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)

	if err != nil {
		return []Comment{}, err
	} else if user_role == NoRole {
		return []Comment{}, custom_errors.ErrAccessDenied
	}

	rows, err := DB.Query(`
		SELECT *
		FROM Comment
		WHERE task_id = $1;`,
		task_id)

	if err != nil {
		log.Println("Error:", err)
		return []Comment{}, custom_errors.ErrDatabaseFailure
	}

	var comments []Comment

	for rows.Next() {
		var comment Comment

		if err := rows.Scan(&comment.ID, &comment.TaskID, &comment.UserID, &comment.Text); err != nil {
			log.Println("Error:", err)
			return []Comment{}, custom_errors.ErrDatabaseFailure
		} else {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

func AddComment(requester_user_id uint, task_id uint, workspace_id uint, text string) (Comment, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)

	if err != nil {
		return Comment{}, err
	} else if user_role == NoRole {
		return Comment{}, custom_errors.ErrAccessDenied
	}

	var comment Comment

	err = DB.QueryRow(`
		INSERT INTO
		Comment(task_id, user_id, text) 
		VALUES($1, $2, $3)
		RETURNING *;`,
		task_id, requester_user_id, text).
		Scan(&comment.ID, &comment.TaskID, &comment.UserID, &comment.Text)

	if err != nil {
		log.Println("Error:", err)
		return Comment{}, custom_errors.ErrDatabaseFailure
	}

	return comment, nil
}

func GetAssociatedUsersWithTask(task_id uint) ([]uint, error) {
	rows, err := DB.Query(`
		SELECT B.user_id
		FROM Task A JOIN UserWorkspaceRole B
		ON A.workspace_id = B.workspace_id
		WHERE A.id = $1;`,
		task_id)

	if err != nil {
		log.Println("Error:", err)
		return []uint{}, custom_errors.ErrDatabaseFailure
	}

	var associated_users []uint

	for rows.Next() {
		var associated_user uint

		if err := rows.Scan(&associated_user); err != nil {
			log.Println("Error:", err)
			return []uint{}, custom_errors.ErrDatabaseFailure
		} else {
			associated_users = append(associated_users, associated_user)
		}
	}

	return associated_users, nil
}

func GetAssociatedUsersWithUser(user_id uint) ([]uint, error) {
	rows, err := DB.Query(`
		SELECT B.user_id
		FROM UserWorkspaceRole A JOIN UserWorkspaceRole B
		ON A.workspace_id = B.workspace_id
		WHERE A.user_id = $1;`,
		user_id)

	if err != nil {
		log.Println("Error:", err)
		return []uint{}, custom_errors.ErrDatabaseFailure
	}

	var associated_users []uint

	for rows.Next() {
		var associated_user uint

		if err := rows.Scan(&associated_user); err != nil {
			log.Println("Error:", err)
			return []uint{}, custom_errors.ErrDatabaseFailure
		} else {
			associated_users = append(associated_users, associated_user)
		}
	}

	return associated_users, nil
}

func GetWorkspaceMembers(workspace_id uint) ([]uint, error) {
	rows, err := DB.Query(`
		SELECT user_id
		FROM UserWorkspaceRole
		WHERE workspace_id = $1;`,
		workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return []uint{}, custom_errors.ErrDatabaseFailure
	}

	var members []uint

	for rows.Next() {
		var member uint

		if err := rows.Scan(&member); err != nil {
			log.Println("Error:", err)
			return []uint{}, custom_errors.ErrDatabaseFailure
		} else {
			members = append(members, member)
		}
	}

	return members, nil
}

func GetWatch(requester_user_id uint, task_id uint, workspace_id uint) (string, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)

	if err != nil {
		return "", err
	} else if user_role == NoRole {
		return "", custom_errors.ErrAccessDenied
	}

	rows, err := DB.Query(`
		SELECT *
		FROM Watch
		WHERE task_id = $1 AND user_id = $2;`,
		task_id, requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return "", custom_errors.ErrDatabaseFailure
	}

	for rows.Next() {
		var member Watch

		if err := rows.Scan(&member); err != nil {
			return "Yes", nil
		}
	}

	return "No", nil
}

func AddWatch(requester_user_id uint, task_id uint, workspace_id uint) error {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)

	if err != nil {
		return err
	} else if user_role == NoRole {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		INSERT INTO
		Watch(task_id, user_id) 
		VALUES($1, $2);`,
		task_id, requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func DeleteWatch(requester_user_id uint, task_id uint, workspace_id uint) error {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)

	if err != nil {
		return err
	} else if user_role == NoRole {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		DELETE FROM Watch
		WHERE task_id = $1 and user_id = $2;`,
		task_id, requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func GetWatchers(task_id uint) ([]uint, error) {
	rows, err := DB.Query(`
		SELECT user_id
		FROM Watch
		WHERE task_id = $1;`,
		task_id)

	if err != nil {
		log.Println("Error:", err)
		return []uint{}, custom_errors.ErrDatabaseFailure
	}

	var watchers []uint

	for rows.Next() {
		var watcher uint

		if err := rows.Scan(&watcher); err != nil {
			log.Println("Error:", err)
			return []uint{}, custom_errors.ErrDatabaseFailure
		} else {
			watchers = append(watchers, watcher)
		}
	}

	return watchers, nil
}
