package user_log

import (
	"time"
	"upsilon_cities_go/lib/db"
)

const (
	UL_Info = 0
	UL_Warn = 1
	UL_Bad  = 2
	UL_Good = 3
)

type UserLog struct {
	ID           int
	UserID       int
	Message      string
	Gravity      int
	Inserted     time.Time
	Acknowledged bool
}

//NewFromCorp register a new Log for corporation owner.
func NewFromCorp(corpID int, gravity int, message string) {
	var ul UserLog
	ul.Gravity = gravity
	ul.Message = message

	dbh := db.New()
	defer dbh.Close()
	ul.InsertFromCorp(dbh, corpID)
}

//New register a new Log for user.
func New(userID int, gravity int, message string) {
	var ul UserLog
	ul.Gravity = gravity
	ul.Message = message
	ul.UserID = userID

	dbh := db.New()
	defer dbh.Close()
	ul.Insert(dbh)
}

//DateStr return string version of inserted date.
func (ul UserLog) DateStr() string {
	return ul.Inserted.Format(time.RFC3339)
}

//GravityStr returns string version of gravity.
func (ul UserLog) GravityStr() string {
	switch ul.Gravity {
	case UL_Info:
		return "Info"
	case UL_Warn:
		return "Warning"
	case UL_Bad:
		return "Problem"
	case UL_Good:
		return "Success"
	default:
		return "Unknown"
	}
}

//MessageStr return a nicer version of Message. but currently does nothing (may include highlighting, html stuff and such, later ;)
func (ul UserLog) MessageStr() string {
	return ul.Message
}
