package corporation

import (
	"encoding/json"
	"upsilon_cities_go/lib/db"
)

//Insert corporation in database.
func (corp *Corporation) Insert(dbh *db.Handler) (err error) {
	if corp.ID != 0 {
		return corp.Update(dbh)
	}

	data, err := corp.dbjsonify()
	rows := dbh.Query("insert into corporations(map_id, name, data) values ($1,$2,$3) returning corporation_id;", corp.GridID, corp.Name, data)
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
	dbh.Query("update corporations set name=$1, data=$2 where corporation_id=$3;", corp.Name, data, corp.ID).Close()

	return
}

//Drop corporation from database
func (corp *Corporation) Drop(dbh *db.Handler) (err error) {

	dbh.Query("delete from corporations where corporation_id=$1", corp.ID).Close()

	corp.ID = 0

	return
}

//ByID Fetch corporation by id; wont link to cities ...
func ByID(dbh *db.Handler, id int) (corp *Corporation, err error) {

	corp = new(Corporation)
	corp.ID = id
	var data []byte
	rows := dbh.Query("select map_id, name, data from corporations where corporation_id=$1;", id)
	for rows.Next() {
		rows.Scan(&corp.GridID, &corp.Name, &data)
	}
	corp.dbunjsonify(data)
	rows.Close()

	rows = dbh.Query("select city_id from cities where corporation_id=$1;", id)
	for rows.Next() {
		var cid int
		rows.Scan(&cid)
		corp.CitiesID = append(corp.CitiesID, cid)
	}
	rows.Close()
	return
}

//ByMapID fetches all corporation related to a map
func ByMapID(dbh *db.Handler, id int) (corps []*Corporation, err error) {

	rows := dbh.Query("select corporation_id, map_id, name, data from corporations where map_id=$1;", id)
	for rows.Next() {
		corp := new(Corporation)
		var data []byte
		rows.Scan(&corp.ID, &corp.GridID, &corp.Name, &data)
		corp.dbunjsonify(data)

		subrow := dbh.Query("select city_id from cities where corporation_id=$1;", id)
		for subrow.Next() {
			var cid int
			subrow.Scan(&cid)
			corp.CitiesID = append(corp.CitiesID, cid)
		}
		subrow.Close()
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
