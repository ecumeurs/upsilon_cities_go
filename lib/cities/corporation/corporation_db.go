package corporation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"upsilon_cities_go/lib/cities/user"
	"upsilon_cities_go/lib/db"
)

//Reload corporation
func (corp *Corporation) Reload(dbh *db.Handler) {
	id := corp.ID
	var data []byte
	rows, err := dbh.Query("select map_id, name, data, (case when user_id is NULL then 0 else user_id end) from corporations where corporation_id=$1;", id)
	if err != nil {
		log.Fatalf("Corporation DB : Failed to select corporation for Reload : %s ", err)
	}

	for rows.Next() {
		rows.Scan(&corp.MapID, &corp.Name, &data, &corp.OwnerID)
	}
	corp.dbunjsonify(data)
	rows.Close()

	rows, err = dbh.Query("select city_id from cities where corporation_id=$1;", id)
	if err != nil {
		log.Fatalf("Corporation DB : Failed to select corporation city for Reload : %s ", err)
	}
	for rows.Next() {
		var cid int
		rows.Scan(&cid)
		corp.CitiesID = append(corp.CitiesID, cid)
	}
	rows.Close()

	rows, err = dbh.Query("select caravan_id from caravans where origin_corporation_id=$1 or target_corporation_id=$2;", id, id)
	if err != nil {
		log.Fatalf("Corporation DB : Failed to select corporation caravan for Reload : %s ", err)
	}
	for rows.Next() {
		var cid int
		rows.Scan(&cid)
		corp.CaravanID = append(corp.CaravanID, cid)
	}
	rows.Close()

}

//Insert corporation in database.
func (corp *Corporation) Insert(dbh *db.Handler) (err error) {
	if corp.ID != 0 {
		return corp.Update(dbh)
	}

	data, err := corp.dbjsonify()

	rows, tmperr := dbh.Query("insert into corporations(map_id, name, data) values ($1,$2,$3) returning corporation_id;", corp.MapID, corp.Name, data)
	if tmperr != nil {
		return fmt.Errorf("Corporation DB : Failed to Insert corporation in database: %s", tmperr)
	}

	for rows.Next() {
		rows.Scan(&corp.ID)
	}
	rows.Close()

	return
}

//Update corporation in database
func (corp *Corporation) Update(dbh *db.Handler) (err error) {
	if corp.ID == 0 {
		return corp.Insert(dbh)
	}

	data, err := corp.dbjsonify()
	if corp.OwnerID == 0 {
		query, tmperr := dbh.Query("update corporations set name=$1, data=$2, user_id=NULL where corporation_id=$3;", corp.Name, data, corp.ID)
		if tmperr != nil {
			return fmt.Errorf("Corporation DB : Failed to update corporation in database: %s", tmperr)
		}
		query.Close()
	} else {
		query, tmperr := dbh.Query("update corporations set name=$1, data=$2, user_id=$3 where corporation_id=$4;", corp.Name, data, corp.OwnerID, corp.ID)
		if tmperr != nil {
			return fmt.Errorf("Corporation DB : Failed to update corporation in database: %s", tmperr)
		}
		query.Close()
	}

	return
}

//Drop corporation from database
func (corp *Corporation) Drop(dbh *db.Handler) (err error) {

	query, err := dbh.Query("delete from corporations where corporation_id=$1", corp.ID)
	if err != nil {
		return fmt.Errorf("Corporation DB : Failed to Drop corporation from database: %s", err)
	}
	query.Close()

	corp.ID = 0

	return
}

//ByID Fetch corporation by id; wont link to cities ...
func ByID(dbh *db.Handler, id int) (corp *Corporation, err error) {

	corp = new(Corporation)
	corp.ID = id
	var data []byte
	rows, err := dbh.Query("select map_id, name, data, (case when user_id is NULL then 0 else user_id end) from corporations where corporation_id=$1;", id)
	if err != nil {
		return nil, fmt.Errorf("Corporation DB : Failed to select corporation (ById): %s", err)
	}
	for rows.Next() {
		rows.Scan(&corp.MapID, &corp.Name, &data, &corp.OwnerID)
	}
	corp.dbunjsonify(data)
	rows.Close()

	rows, err = dbh.Query("select city_id from cities where corporation_id=$1;", id)
	if err != nil {
		return corp, fmt.Errorf("Corporation DB : Failed to select cities (ById): %s", err)
	}
	for rows.Next() {
		var cid int
		rows.Scan(&cid)
		corp.CitiesID = append(corp.CitiesID, cid)
	}
	rows.Close()

	rows, err = dbh.Query("select caravan_id from caravans where origin_corporation_id=$1 or target_corporation_id=$2;", id, id)
	if err != nil {
		return corp, fmt.Errorf("Corporation DB : Failed to select  caravan (ByID): %s", err)
	}
	for rows.Next() {
		var cid int
		rows.Scan(&cid)
		corp.CaravanID = append(corp.CaravanID, cid)
	}
	rows.Close()
	return
}

//ByMapID fetches all corporation related to a map
func ByMapID(dbh *db.Handler, id int) (corps []*Corporation, err error) {

	rows, err := dbh.Query("select corporation_id, map_id, name, data, (case when user_id is NULL then 0 else user_id end)  from corporations where map_id=$1;", id)
	if err != nil {
		return nil, fmt.Errorf("Corporation DB : Failed to select corporation (ByMapID): %s", err)
	}

	for rows.Next() {
		corp := new(Corporation)
		var data []byte
		rows.Scan(&corp.ID, &corp.MapID, &corp.Name, &data, &corp.OwnerID)
		corp.dbunjsonify(data)

		subrow, err := dbh.Query("select city_id from cities where corporation_id=$1;", corp.ID)
		if err != nil {
			return corps, fmt.Errorf("Corporation DB : Failed to select cities (ByMapID): %s", err)
		}
		for subrow.Next() {
			var cid int
			subrow.Scan(&cid)
			corp.CitiesID = append(corp.CitiesID, cid)
		}
		subrow.Close()

		subrow, err = dbh.Query("select caravan_id from caravans where origin_corporation_id=$1 or target_corporation_id=$2;", corp.ID, corp.ID)
		if err != nil {
			return corps, fmt.Errorf("Corporation DB : Failed to select caravan (ByMapID): %s", err)
		}
		for subrow.Next() {
			var cid int
			subrow.Scan(&cid)
			corp.CaravanID = append(corp.CaravanID, cid)
		}
		subrow.Close()

		corps = append(corps, corp)
	}
	rows.Close()

	return
}

//ByMapIDByUserID fetches all corporation related to a map by id, may not
func ByMapIDByUserID(dbh *db.Handler, id int, userID int) (corp *Corporation, err error) {

	rows, err := dbh.Query("select corporation_id, map_id, name, data, (case when user_id is NULL then 0 else user_id end)  from corporations where map_id=$1 and user_id=$2;", id, userID)
	if err != nil {
		return nil, fmt.Errorf("Corporation DB : Failed to select corporation (ByMapIDbyUserID) : %s", err)
	}
	for rows.Next() {
		corp := new(Corporation)
		var data []byte
		rows.Scan(&corp.ID, &corp.MapID, &corp.Name, &data, &corp.OwnerID)
		corp.dbunjsonify(data)

		subrow, err := dbh.Query("select city_id from cities where corporation_id=$1;", corp.ID)
		if err != nil {
			return corp, fmt.Errorf("Corporation DB : Failed to select cities (ByMapIDbyUserID) : %s", err)
		}
		for subrow.Next() {
			var cid int
			subrow.Scan(&cid)
			corp.CitiesID = append(corp.CitiesID, cid)
		}
		subrow.Close()

		subrow, err = dbh.Query("select caravan_id from caravans where origin_corporation_id=$1 or target_corporation_id=$2;", corp.ID, corp.ID)
		if err != nil {
			return corp, fmt.Errorf("Corporation DB : Failed to select caravan (ByMapIDbyUserID) : %s", err)
		}
		for subrow.Next() {
			var cid int
			subrow.Scan(&cid)
			corp.CaravanID = append(corp.CaravanID, cid)
		}
		subrow.Close()
		rows.Close()
		return corp, nil
	}

	return nil, errors.New("no corporation matching found")
}

//ByMapIDClaimable fetches all corporation related to a map thar don't have owner
func ByMapIDClaimable(dbh *db.Handler, id int) (corps []*Corporation, err error) {

	rows, err := dbh.Query("select corporation_id, map_id, name, data from corporations where map_id=$1 and user_id is NULL;", id)
	if err != nil {
		return nil, fmt.Errorf("Corporation DB : Failed to select corporation (ByMapIDClaimable): %s", err)
	}
	for rows.Next() {
		corp := new(Corporation)
		var data []byte
		rows.Scan(&corp.ID, &corp.MapID, &corp.Name, &data)
		corp.dbunjsonify(data)

		subrow, err := dbh.Query("select city_id from cities where corporation_id=$1;", corp.ID)
		if err != nil {
			return corps, fmt.Errorf("Corporation DB : Failed to select corporation cities (ByMapIDClaimable): %s", err)
		}
		for subrow.Next() {
			var cid int
			subrow.Scan(&cid)
			corp.CitiesID = append(corp.CitiesID, cid)
		}
		subrow.Close()

		subrow, err = dbh.Query("select caravan_id from caravans where origin_corporation_id=$1 or target_corporation_id=$2;", corp.ID, corp.ID)
		if err != nil {
			return corps, fmt.Errorf("Corporation DB : Failed to select corportation caravan (ByMapIDClaimable): %s", err)
		}
		for subrow.Next() {
			var cid int
			subrow.Scan(&cid)
			corp.CaravanID = append(corp.CaravanID, cid)
		}
		subrow.Close()

		corps = append(corps, corp)
	}
	rows.Close()

	return
}

type dbCorporation struct {
}

func (corp *Corporation) dbjsonify() (res []byte, err error) {
	var tmp dbCorporation
	return json.Marshal(tmp)
}

func (corp *Corporation) dbunjsonify(fromJSON []byte) (err error) {
	var db dbCorporation
	err = json.Unmarshal(fromJSON, &db)
	if err != nil {
		return err
	}

	return nil
}

//Claim by user a corporation.
func Claim(dbh *db.Handler, usr *user.User, corp *Corporation) error {
	if corp.OwnerID != 0 {
		return errors.New("unable to claim corporation as already owned by someone else")
	}

	corp.OwnerID = usr.ID
	return corp.Update(dbh)
}
