package user

import "time"

type User struct {
	ID        int
	Login     string
	Email     string
	Password  string
	LastLogin time.Time

	// admin

	NeedNewPassword bool
	Enabled         bool
}
