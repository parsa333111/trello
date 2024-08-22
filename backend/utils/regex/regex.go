package regex_utils

import (
	"log"

	"github.com/dlclark/regexp2"
)

const (
	usernameRegex = "^[A-Za-z0-9]{4,12}$"
	emailRegex    = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*$"
	passwordRegex = "^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[@#$!%*?&])[A-Za-z0-9@#$!%*?&]{8,32}$"
)

func ValidateUsername(username string) bool {
	regex := regexp2.MustCompile(usernameRegex, 0)
	match, err := regex.MatchString(username)
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	return match
}

func ValidateEmail(email string) bool {
	regex := regexp2.MustCompile(emailRegex, 0)
	match, err := regex.MatchString(email)
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	return match
}

func ValidatePassword(password string) bool {
	regex := regexp2.MustCompile(passwordRegex, 0)
	match, err := regex.MatchString(password)
	if err != nil {
		log.Println("Error:", err)
		return false
	}
	return match
}
