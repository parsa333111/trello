package database

import (
	"fmt"
	"log"
	"time"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

func GetAssignedTasks(requester_user_id uint) ([]Task, error) {
	rows, err := DB.Query(`
		SELECT * 
		FROM Task
		WHERE assignee_id = $1;`,
		requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return []Task{}, custom_errors.ErrDatabaseFailure
	}

	var tasks []Task

	for rows.Next() {
		var task Task

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.EstimatedTime,
			&task.ActualTime,
			&task.DueDate,
			&task.Priority,
			&task.WorkspaceID,
			&task.AssigneeID,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ImageURL,
		); err != nil {
			log.Println("Error:", err)
			return []Task{}, custom_errors.ErrDatabaseFailure
		} else {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

func GetAllTasksInWorkspace(requester_user_id uint, workspace_id uint) ([]Task, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return []Task{}, err
	} else if user_role == NoRole {
		return []Task{}, custom_errors.ErrAccessDenied
	}

	rows, err := DB.Query(`
		SELECT * 
		FROM Task
		WHERE workspace_id = $1;`,
		workspace_id)

	if err != nil {
		log.Println("Error:", err)
		return []Task{}, custom_errors.ErrDatabaseFailure
	}

	var tasks []Task

	for rows.Next() {
		var task Task

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.EstimatedTime,
			&task.ActualTime,
			&task.DueDate,
			&task.Priority,
			&task.WorkspaceID,
			&task.AssigneeID,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ImageURL,
		); err != nil {
			log.Println("Error:", err)
			return []Task{}, custom_errors.ErrDatabaseFailure
		} else {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

func CreateTaskInWorkspace(requester_user_id uint, workspace_id uint, title string, description string, estimatedtime int, actualtime int, duedate time.Time, priority int, assigneeID uint, imageURL string) (Task, error) {
	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return Task{}, err
	} else if user_role != Admin && user_role != Owner {
		return Task{}, custom_errors.ErrAccessDenied
	}

	assignee_user_role, err := getUserWorkspaceRole(assigneeID, workspace_id)
	if err != nil {
		return Task{}, err
	} else if assignee_user_role == NoRole {
		return Task{}, custom_errors.ErrInvalidArguments
	}

	if ok := checkDuplicateTaskTitle(workspace_id, title); !ok {
		return Task{}, custom_errors.ErrDuplicateTaskTitle
	}

	var task Task

	err = DB.QueryRow(`
		INSERT INTO
		Task(title,description,status,estimated_time,actual_time,due_date,priority,workspace_id,assignee_id,created_at,updated_at,image_url)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING *;`,
		title, description, Planned,
		estimatedtime, actualtime, duedate,
		priority, workspace_id, assigneeID,
		time.Now(), time.Now(), imageURL).
		Scan(&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.EstimatedTime,
			&task.ActualTime,
			&task.DueDate,
			&task.Priority,
			&task.WorkspaceID,
			&task.AssigneeID,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ImageURL)

	if nil != err {
		log.Println("Error:", err)
		return Task{}, custom_errors.ErrDatabaseFailure
	}

	return task, nil
}

func GetDetailsOfTask(requester_user_id uint, workspace_id uint, task_id uint) (Task, error) {
	actual_workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		log.Println("Error:", err)
		return Task{}, custom_errors.ErrDatabaseFailure
	} else if actual_workspace_id != workspace_id {
		return Task{}, custom_errors.ErrInvalidArguments
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return Task{}, err
	} else if user_role == NoRole {
		return Task{}, custom_errors.ErrAccessDenied
	}

	var task Task

	err = DB.QueryRow(`
		SELECT * 
		FROM Task
		WHERE id = $1;`,
		task_id).
		Scan(&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.EstimatedTime,
			&task.ActualTime,
			&task.DueDate,
			&task.Priority,
			&task.WorkspaceID,
			&task.AssigneeID,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ImageURL)

	if err != nil {
		log.Println("Error:", err)
		return Task{}, custom_errors.ErrDatabaseFailure
	}

	return task, nil
}

func UpdateStatusOfTask(requester_user_id uint, workspace_id uint, task_id uint, status string) error {
	actual_workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	} else if actual_workspace_id != workspace_id {
		return custom_errors.ErrInvalidArguments
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		UPDATE Task 
		SET status = $1
		WHERE id = $2;`,
		status, task_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func UpdateDetailsOfTask(requester_user_id uint, workspace_id uint, task_id uint, title string, description string, actualtime int, duedate time.Time, priority int, assigneeID uint, imageURL string) error {
	actual_workspace_id, err := getTaskWorkspaceID(task_id)
	fmt.Print("h1", err)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	} else if actual_workspace_id != workspace_id {
		return custom_errors.ErrInvalidArguments
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	fmt.Print("h2", err)

	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	assignee_user_role, err := getUserWorkspaceRole(assigneeID, workspace_id)
	fmt.Print("h3", err)
	if err != nil {
		return err
	} else if assignee_user_role == NoRole {
		return custom_errors.ErrInvalidArguments
	}

	if ok := checkDuplicateTaskTitleWithoutSelf(workspace_id, title, task_id); !ok {
		return custom_errors.ErrDuplicateTaskTitle
	}

	_, err = DB.Exec(`
		UPDATE Task 
		SET title = $1, description = $2, actual_time = $3, due_date = $4, priority = $5, updated_at = $6, assignee_id = $7, image_url = $8
		WHERE id = $9;`,
		title, description, actualtime, duedate,
		priority, time.Now(), assigneeID, imageURL, task_id)
	fmt.Print("h4", err)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func DeleteTask(requester_user_id uint, workspace_id uint, task_id uint) error {
	actual_workspace_id, err := getTaskWorkspaceID(task_id)
	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	} else if actual_workspace_id != workspace_id {
		return custom_errors.ErrInvalidArguments
	}

	user_role, err := getUserWorkspaceRole(requester_user_id, workspace_id)
	if err != nil {
		return err
	} else if user_role != Admin && user_role != Owner {
		return custom_errors.ErrAccessDenied
	}

	_, err = DB.Exec(`
		DELETE FROM Task
		WHERE id = $1;`,
		task_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}
