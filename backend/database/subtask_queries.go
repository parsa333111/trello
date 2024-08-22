package database

import (
	"log"
	"time"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

func GetAllSubtasksInTask(requester_user_id uint, task_id uint) ([]Subtask, error) {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return []Subtask{}, custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return []Subtask{}, err
	} else if user_role == NoRole {
		return []Subtask{}, custom_errors.ErrAccessDenied
	}

	rows, err := DB.Query(`
		SELECT * 
		FROM Subtask
		WHERE task_id = $1;`,
		task_id)

	if err != nil {
		log.Println("Error:", err)
		return []Subtask{}, custom_errors.ErrDatabaseFailure
	}

	var subtasks []Subtask

	for rows.Next() {
		var subtask Subtask

		if err := rows.Scan(
			&subtask.ID,
			&subtask.TaskID,
			&subtask.Title,
			&subtask.IsCompleted,
			&subtask.AssigneeID,
			&subtask.CreatedAt,
			&subtask.UpdatedAt); err != nil {
			log.Println("Error:", err)
			return []Subtask{}, custom_errors.ErrDatabaseFailure
		} else {
			subtasks = append(subtasks, subtask)
		}
	}

	return subtasks, nil
}

func CreateSubtaskInTask(requester_user_id uint, task_id uint, title string, assignee_id uint) (Subtask, error) {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return Subtask{}, custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return Subtask{}, err
	} else if user_role != Admin && user_role != Owner {
		return Subtask{}, custom_errors.ErrAccessDenied
	}

	if ok := checkDuplicateSubtaskTitle(task_id, title); !ok {
		return Subtask{}, custom_errors.ErrDuplicateSubtaskTitle
	}

	var subtask Subtask

	err = DB.QueryRow(`
		INSERT INTO
		Subtask(task_id,title,is_completed,assignee_id,created_at,updated_at)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING *;`,
		task_id, title, No, assignee_id,
		time.Now(), time.Now()).
		Scan(&subtask.ID,
			&subtask.TaskID,
			&subtask.Title,
			&subtask.IsCompleted,
			&subtask.AssigneeID,
			&subtask.CreatedAt,
			&subtask.UpdatedAt)

	if nil != err {
		log.Println("Error:", err)
		return Subtask{}, custom_errors.ErrDatabaseFailure
	}

	return subtask, nil
}

func GetDetailsOfSubtask(requester_user_id uint, task_id uint, subtask_id uint) (Subtask, error) {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return Subtask{}, custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return Subtask{}, err
	} else if user_role == NoRole {
		return Subtask{}, custom_errors.ErrAccessDenied
	}

	var subtask Subtask

	err = DB.QueryRow(`
		SELECT * 
		FROM Subtask
		WHERE id = $1;`,
		subtask_id).
		Scan(&subtask.ID,
			&subtask.TaskID,
			&subtask.Title,
			&subtask.IsCompleted,
			&subtask.AssigneeID,
			&subtask.CreatedAt,
			&subtask.UpdatedAt)

	if err != nil {
		log.Println("Error:", err)
		return Subtask{}, custom_errors.ErrDatabaseFailure
	}

	return subtask, nil
}

func UpdateDetailsOfSubtask(requester_user_id uint, task_id uint, subtask_id uint, title string, is_completed string, assignee_id uint) error {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	if ok := checkDuplicateSubtaskTitle(task_id, title); !ok {
		return custom_errors.ErrDuplicateSubtaskTitle
	}

	_, err = DB.Exec(`
		UPDATE Subtask 
		SET title = $1, is_completed = $2, assignee_id = $3, updated_at = $4
		WHERE id = $5;`,
		title, is_completed, assignee_id,
		time.Now(), subtask_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}
func UpdateDetailsOfSubtaskAssigneeID(requester_user_id uint, task_id uint, subtask_id uint, assignee_id uint) error {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		UPDATE Subtask 
		SET assignee_id = $1, updated_at = $2
		WHERE id = $3;`,
		assignee_id, time.Now(), subtask_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func UpdateDetailsOfSubtaskTitle(requester_user_id uint, task_id uint, subtask_id uint, title string) error {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	if ok := checkDuplicateSubtaskTitle(task_id, title); !ok {
		return custom_errors.ErrDuplicateSubtaskTitle
	}

	_, err = DB.Exec(`
		UPDATE Subtask 
		SET title = $1, updated_at = $2
		WHERE id = $3;`,
		title, time.Now(), subtask_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func UpdateDetailsOfSubtaskStatus(requester_user_id uint, task_id uint, subtask_id uint, is_completed string) error {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		UPDATE Subtask 
		SET is_completed = $1, updated_at = $2
		WHERE id = $3;`,
		is_completed, time.Now(), subtask_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func DeleteSubtask(requester_user_id uint, task_id uint, subtask_id uint) error {
	workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		return custom_errors.ErrDatabaseFailure
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		DELETE FROM Subtask
		WHERE id = $1;;`,
		subtask_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}
