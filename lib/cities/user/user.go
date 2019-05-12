package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Login     string
	Email     string
	Password  string
	LastLogin time.Time

	// admin

	NeedNewPassword bool
	Enabled         bool
	Admin           bool
}

// courtesy to https://gowebexamples.com/password-hashing/

//HashPassword generate a hash based on nice password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

//CheckPasswordHash will check provided password against hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
