package user_log

import (
	"errors"
	"time"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/db"

	"github.com/lib/pq"
)

//Insert add log to database
func (ul UserLog) Insert(dbh *db.Handler) error {
	if ul.UserID == 0 {
		return errors.New("can't store user log without an user_id")
	}

	rows, _ := dbh.Query("insert into user_logs(user_id, message, gravity) values($1,$2,$3) returning user_log_id;", ul.UserID, ul.Message, ul.Gravity)
	for rows.Next() {
		rows.Scan(&ul.ID)
	}

	return nil
}

//InsertFromCorp add log to database and seek user from database.
func (ul UserLog) InsertFromCorp(dbh *db.Handler, corpID int) error {
	if corpID == 0 {
		return errors.New("can't store user log without an corporation_id")
	}

	crp, err := corporation_manager.GetCorporationHandler(corpID)
	if err != nil {
		return errors.New("unknown corporation")
	}

	if crp.Get().OwnerID == 0 {
		return errors.New("corporation without owner")
	}

	ul.UserID = crp.Get().OwnerID
	return ul.Insert(dbh)
}

//MarkRead mark message as displayed
func MarkRead(dbh *db.Handler, ids []int) error {
	if len(ids) == 0 {
		return errors.New("can't store user log without an user_log_id")
	}

	query, _ := dbh.Query("update user_logs set aknowledged = (now() at time zone 'utc') where user_log_id=ANY($1)", pq.Array(ids))
	query.Close()
	return nil
}

//LastMessages get last unacknowledged messages
func LastMessages(dbh *db.Handler, userID int) (res []UserLog) {

	rows, _ := dbh.Query("select user_log_id, user_id, message, gravity, inserted, acknowledged != NULL as ack from user_logs where user_id=$1 order by user_log_id desc", userID)
	for rows.Next() {
		var ul UserLog
		rows.Scan(&ul.ID, &ul.UserID, &ul.Message, &ul.Gravity, &ul.Inserted, &ul.Acknowledged)
		res = append(res, ul)
	}

	return res
}

//Since get last unacknowledged messages
func Since(dbh *db.Handler, userID int, date time.Time) (res []UserLog) {

	rows, _ := dbh.Query("select user_log_id, user_id, message, gravity, inserted, acknowledged != NULL as ack from user_logs where user_id=$1 and inserted > $2 order by user_log_id desc", userID, date)
	for rows.Next() {
		var ul UserLog
		rows.Scan(&ul.ID, &ul.UserID, &ul.Message, &ul.Gravity, &ul.Inserted, &ul.Acknowledged)
		res = append(res, ul)
	}

	return res
}
