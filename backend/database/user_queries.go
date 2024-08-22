package database

import (
	"log"
	"time"

	"github.com/skye-tan/trello/backend/utils/custom_errors"
)

func GetUsers() ([]User, error) {
	rows, err := DB.Query(`SELECT * FROM Users;`)

	if err != nil {
		log.Println("Error:", err)
		return []User{}, custom_errors.ErrDatabaseFailure
	}

	var users []User

	for rows.Next() {
		var user User

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt); err != nil {
			log.Println("Error:", err)
			return []User{}, custom_errors.ErrDatabaseFailure
		} else {
			users = append(users, user)
		}
	}

	return users, nil
}

func CreateUser(username string, email string, password_hash []byte) error {
	if ok := checkDuplicateEmail(email); !ok {
		return custom_errors.ErrDuplicateEmail
	}
	if ok := checkDuplicateUsername(username); !ok {
		return custom_errors.ErrDuplicateUsername
	}

	_, err := DB.Exec(`
		INSERT INTO
		Users(username, email, password_hash, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5);`,
		username, email, password_hash,
		time.Now(), time.Now())

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func GetUserByID(user_id uint) (User, error) {
	var user User
	err := DB.QueryRow(`
		SELECT * FROM Users
		WHERE id = $1;`,
		user_id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt)

	if err != nil {
		log.Println("Error:", err)
		return User{}, custom_errors.ErrDatabaseFailure
	}

	return user, nil
}

func GetUserByUsername(username string) (User, error) {
	var user User
	err := DB.QueryRow(`
		SELECT * FROM Users
		WHERE username = $1;`,
		username).
		Scan(&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt)

	if err != nil {
		log.Println("Error:", err)
		return User{}, custom_errors.ErrDatabaseFailure
	}

	return user, nil
}

func UpdateUserUsername(requester_user_id uint, username string) error {
	if ok := checkDuplicateUsername(username); !ok {
		return custom_errors.ErrDuplicateUsername
	}

	_, err := DB.Exec(`
		UPDATE Users
		SET username = $1, updated_at = $2
		WHERE id = $3;`,
		username, time.Now(), requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func UpdateUserPassword(requester_user_id uint, password_hash []byte) error {
	_, err := DB.Exec(`
		UPDATE Users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3;`,
		password_hash, time.Now(), requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}

func DeleteUser(requester_user_id uint) error {
	_, err := DB.Exec(`
		DELETE FROM Users
		WHERE id = $1;`,
		requester_user_id)

	if err != nil {
		log.Println("Error:", err)
		return custom_errors.ErrDatabaseFailure
	}

	return nil
}
