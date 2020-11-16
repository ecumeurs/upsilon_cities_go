package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"upsilon_cities_go/lib/db"
)

//CheckMailAvailability returns true when email are unknown
func CheckMailAvailability(dbh *db.Handler, email string) (bool, error) {
	rows, err := dbh.Query("select count(*) from users where email=$1", email)
	if err != nil {
		return false, fmt.Errorf("User DB: CheckMailAvailability Failed to Select . %s", err)
	}
	rows.Next()
	nb := 0
	rows.Scan(&nb)

	rows.Close()
	return nb == 0, nil
}

//CheckLoginAvailability returns true when login are unknown
func CheckLoginAvailability(dbh *db.Handler, login string) (bool, error) {
	rows, err := dbh.Query("select count(*) from users where login=$1", login)
	if err != nil {
		return false, fmt.Errorf("User DB: CheckLoginAvailability Failed to Select . %s", err)
	}
	rows.Next()
	nb := 0
	rows.Scan(&nb)

	rows.Close()
	return nb == 0, nil
}

//Insert user in database
func (user *User) Insert(dbh *db.Handler) error {

	rows, err := dbh.Query("insert into users(login) values($1) returning user_id", user.Login)
	if err != nil {
		return fmt.Errorf("User DB: Failed to Insert. %s", err)
	}

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

	query, err := dbh.Query(`
		update users set 
			login=$1
			, email=$2
			, password=$3
			, enabled=$4
			, admin=$5
			, data=$6
			where user_id=$7`,

		user.Login, user.Email, user.Password, user.Enabled, user.Admin, js, user.ID)
	query.Close()
	if err != nil {
		return fmt.Errorf("User DB: Failed to Update. %s", err)
	}

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

	query, err := dbh.Query(`
		update users set 
			, data=$1
			where user_id=$2`,
		js, user.ID)
	if err != nil {
		return fmt.Errorf("User DB: Failed to ShortUpdate. %s", err)
	}
	query.Close()

	log.Printf("User: Updated user's data %d - %s", user.ID, user.Login)
	return nil
}

//UpdatePassword updates only data from user.
func (user *User) UpdatePassword(dbh *db.Handler) error {
	if user.ID == 0 {
		return user.Insert(dbh)
	}

	user.NeedNewPassword = false

	js, err := user.dbjsonify()
	if err != nil {
		log.Printf("User: Failed to jsonify user data")
		return err
	}

	query, err := dbh.Query(`
		update users set 
			password=$1,
			data=$2
			where user_id=$3`,
		user.Password, js, user.ID)
	if err != nil {
		return fmt.Errorf("User DB: Failed to UpdatePassword. %s", err)
	}
	query.Close()

	log.Printf("User: Updated user's password %d - %s", user.ID, user.Login)
	return nil
}

//LogsIn updates last login date.
func (user *User) LogsIn(dbh *db.Handler, id string) error {
	if user.ID == 0 {
		return errors.New("can't login an unknown user")
	}

	query, err := dbh.Query("update users set last_login=$1, key=$2 where user_id=$3", user.LastLogin, id, user.ID)
	if err != nil {
		return fmt.Errorf("User DB: Failed to Update LogsIn. %s", err)
	}
	query.Close()

	log.Printf("User: Updated Login date of user %d - %s", user.ID, user.Login)
	return nil
}

//Drop user from database
func Drop(dbh *db.Handler, id int) error {
	log.Printf("User: Dropped user %d", id)

	query, err := dbh.Query("delete from http_sessions hs USING users usr where (usr.user_id = $1 AND usr.key = hs.key)", id)
	if err != nil {
		return fmt.Errorf("User DB: Failed to Drop http_sessions. %s", err)
	}
	query.Close()

	query, err = dbh.Query("delete from users where user_id=$1", id)
	if err != nil {
		return fmt.Errorf("User DB: Failed to Drop Users. %s", err)
	}

	query.Close()

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

	rows, err := dbh.Query("select user_id, login, email, password, enabled, admin, last_login, data from users where login=$1", login)
	if err != nil {
		return nil, fmt.Errorf("User DB: Failed to seek user by login (ByLogin) %s", err)
	}
	for rows.Next() {
		user := convert(rows)
		rows.Close()
		return user, nil
	}

	return nil, fmt.Errorf("failed to find requested user %s", login)
}

//ByID seek user by login
func ByID(dbh *db.Handler, id int) (*User, error) {

	rows, err := dbh.Query("select user_id, login, email, password, enabled, admin, last_login, data from users where user_id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("User DB: Failed to seek user by login (ByID) %s", err)
	}
	for rows.Next() {
		user := convert(rows)
		rows.Close()
		return user, nil
	}

	return nil, fmt.Errorf("failed to find requested user %d", id)
}

//All return a listing of all user
func All(dbh *db.Handler) (res []*User) {

	rows, err := dbh.Query("select user_id, login, email, password, enabled, admin, last_login, data from users")
	if err != nil {
		log.Fatalf("User DB: Failed to return a listing of all user (All). %s", err)
	}
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
