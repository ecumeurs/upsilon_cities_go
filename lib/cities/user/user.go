package user

import (
	"time"
	"upsilon_cities_go/config"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Login     string
	Email     string
	Password  string `json:"-"`
	LastLogin time.Time

	// admin

	NeedNewPassword bool
	Enabled         bool
	Admin           bool
}

// courtesy to https://gowebexamples.com/password-hashing/

//New create a new user.
func New() *User {
	usr := new(User)

	usr.NeedNewPassword = false
	usr.Enabled = config.USER_ENABLED_BY_DEFAULT
	usr.Admin = config.USER_ADMIN_BY_DEFAULT
	return usr
}

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

//PrettyLastLogin stringify last login.
func (user *User) PrettyLastLogin() string {
	return user.LastLogin.Format(time.RFC3339)
}
