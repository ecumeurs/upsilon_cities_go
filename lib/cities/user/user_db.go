package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"upsilon_cities_go/lib/db"
)

//CheckAvailability returns true when login/email are unknown
func CheckAvailability(dbh *db.Handler, login, email string) bool {
	rows := dbh.Query("select count(*) from users where login=$1 or email=$2", login, email)

	rows.Next()
	nb := 0
	rows.Scan(nb)

	rows.Close()
	return nb == 0
}

//Insert user in database
func (user *User) Insert(dbh *db.Handler) error {

	rows := dbh.Query("insert into users(login) values($1) returning user_id", user.Login)
	for rows.Next() {
		rows.Scan(&user.ID)
	}
	rows.Close()

	log.Printf("User: Inserted user %d - %s", user.ID, user.Login)
	return user.Update(dbh)
}

//Update user in database
func (user *User) Update(dbh *db.Handler) error {
	if user.ID == 0 {
		return user.Insert(dbh)
	}

	js, err := user.dbjsonify()
	if err != nil {
		log.Printf("User: Failed to jsonify user data")
		return err
	}

	dbh.Query(`
		update users set 
			login=$1
			, email=$2
			, password=$3
			, enabled=$4
			, admin=$5
			, data=$6
			where user_id=$7`,

		user.Login, user.Email, user.Password, user.Enabled, user.Admin, js, user.ID).Close()

	log.Printf("User: Updated user %d - %s", user.ID, user.Login)
	return nil
}

//ShortUpdate updates only data from user.
func (user *User) ShortUpdate(dbh *db.Handler) error {
	if user.ID == 0 {
		return user.Insert(dbh)
	}

	js, err := user.dbjsonify()
	if err != nil {
		log.Printf("User: Failed to jsonify user data")
		return err
	}

	dbh.Query(`
		update users set 
			, data=$1
			where user_id=$2`,
		js, user.ID).Close()

	log.Printf("User: Updated user's data %d - %s", user.ID, user.Login)
	return nil
}

//UpdatePassword updates only data from user.
func (user *User) UpdatePassword(dbh *db.Handler) error {
	if user.ID == 0 {
		return user.Insert(dbh)
	}

	dbh.Query(`
		update users set 
			, password=$1
			where user_id=$2`,
		user.Password, user.ID).Close()

	log.Printf("User: Updated user's password %d - %s", user.ID, user.Login)
	return nil
}

//LogsIn updates last login date.
func (user *User) LogsIn(dbh *db.Handler) error {
	if user.ID == 0 {
		return errors.New("can't login an unknown user")
	}

	dbh.Query("update users set last_login=$1 where user_id=$2", user.LastLogin, user.ID).Close()

	log.Printf("User: Updated Login date of user %d - %s", user.ID, user.Login)
	return nil
}

//Drop user from database
func Drop(dbh *db.Handler, id int) error {
	log.Printf("User: Dropped user %d", id)
	dbh.Query("delete from users where user_id=$1", id).Close()
	return nil
}

func convert(rows *sql.Rows) (usr *User) {

	var js []byte

	usr = new(User)
	rows.Scan(&usr.ID,
		&usr.Login,
		&usr.Email,
		&usr.Password,
		&usr.Enabled,
		&usr.Admin,
		&usr.LastLogin,
		&js)

	usr.dbunjsonify(js)

	return
}

//ByLogin seek user by login
func ByLogin(dbh *db.Handler, login string) (*User, error) {

	rows := dbh.Query("select user_id, login, email, password, enabled, admin, last_login, data from users where login=$1", login)
	for rows.Next() {
		user := convert(rows)
		rows.Close()
		return user, nil
	}

	return nil, fmt.Errorf("failed to find requested user %s", login)
}

//All return a listing of all user
func All(dbh *db.Handler) (res []*User) {

	rows := dbh.Query("select user_id, login, email, password, enabled, admin, last_login, data from users")
	for rows.Next() {
		user := convert(rows)

		res = append(res, user)
	}
	rows.Close()

	return res
}

type dbUser struct {
	NeedNewPassword bool
}

func (user *User) dbjsonify() (res []byte, err error) {
	var tmp dbUser
	tmp.NeedNewPassword = user.NeedNewPassword
	return json.Marshal(tmp)
}

func (user *User) dbunjsonify(fromJSON []byte) (err error) {
	var db dbUser
	err = json.Unmarshal(fromJSON, &db)
	if err != nil {
		return err
	}

	user.NeedNewPassword = db.NeedNewPassword
	return nil
}
