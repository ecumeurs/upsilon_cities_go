package city

import (
	"errors"
	"log"
	"reflect"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/db"

	"github.com/lib/pq"
)

//City
type City struct {
	ID         int
	Location   node.Point
	Neighbours []*City
	Roads      []node.Pathway
}

//Insert Stores or Update city as necessary. Won't insert neighbours at that time ;)
func (city *City) Insert(dbh *db.Handler, mapID int) error {
	if city.ID <= 0 {
		// this is a new city. Simply attribute it a new ID
		res := dbh.Query("insert into cities(city_name, map_id) values('', $1) returning city_id;", mapID)
		for res.Next() {
			res.Scan(&city.ID)
		}

		if city.ID <= 0 {
			log.Fatalln("City: Failed to insert City in database.")
		}
		log.Printf("City: Added City %d: %s to database", city.ID, "")
	} else {
		return errors.New("can't insert an already identified city")
	}
	return nil
}

//Update City, will repsert neighbours as appropriate.
func (city *City) Update(dbh *db.Handler) error {
	if city.ID <= 0 {
		return errors.New("can't update an unknown identified city")
	} else {
		// prepare json dump
		json, err := city.dbjsonify()
		if err != nil {
			return err
		}

		// dhb.db.Query has formater stuff ;)
		res := dbh.Query(`update table cities set 
			updated_at=(now() at time zone 'utc')
			date=$1
			where city_id=$2
		`, json, city.ID)
		for res.Next() {
			res.Scan(&city.ID)
		}

		err = city.dbCheckNeighbours(dbh)
		if err != nil {
			return err
		}
	}
	log.Printf("City: Updated City %d: %s to database", city.ID, "")

	return nil
}

func (city *City) dbCheckNeighbours(dbh *db.Handler) error {
	// select known neighbours
	neighboursRows := dbh.Query("select * from neighbouring_cities where from_city_id=$1", city.ID)
	var neighbours map[int]int
	for neighboursRows.Next() {
		var id int
		neighboursRows.Scan(&id)
		neighbours[id] = id
	}
	// compare with actual neighbours

	missingNeighbours := neighbours
	var newNeighbours []int
	for _, v := range city.Neighbours {
		if _, found := neighbours[v.ID]; found {
			// well it's already there so we don't care
			delete(missingNeighbours, v.ID) // it's not missing ;)
		} else {
			// it's not there, so it's new
			newNeighbours = append(newNeighbours, v.ID)
		}
	}

	// drop disappeared neighbours

	dbh.Query("delete from neighbouring_cities where from_city_id=$1 and to_city_id=ANY($2)", city.ID, pq.Array(reflect.ValueOf(missingNeighbours).MapKeys()))

	// be nice and remove reverse as well ;)

	dbh.Query("delete from neighbouring_cities where to_city_id=$1 and from_city_id=ANY($2)", city.ID, pq.Array(reflect.ValueOf(missingNeighbours).MapKeys()))

	// add missing neighbours

	for _, v := range newNeighbours {
		dbh.Query("insert into neighbouring_cities(to_city_id, from_city_id) values ($1,$2),($3,$4)", city.ID, v, v, city.ID)
	}

	return nil
}

// prepare the json version for database, may not be the appropriate one for API ;)
// Useless for now, but will be usefull when city get storage and stuff like that.
func (city *City) dbjsonify() (res string, err error) {
	err = nil
	res = "{}"
	return
}

// reverse operation unpack json ;)
func (city *City) dbunjsonify(_ string) (err error) {
	err = nil
	return
}

//ByID Fetch a city by id; note, won't load neighbouring cities ... or maybe only their ids ? ...
func ByID(dbh *db.Handler, id int) (city *City, neighbouringCities []int, err error) {
	err = nil

	city = new(City)
	rows := dbh.Query("select city_id from cities where city_id=$1", id)
	for rows.Next() {
		// hopefully there is only one ;) city_id is supposed to be unique.
		// atm only read city_id ;)
		rows.Scan(&city.ID)
	}

	// seek its neighbours
	rows = dbh.Query("select to_city_id from neighbouring_cities where from_city_id=$1", id)
	for rows.Next() {
		var nid int
		rows.Scan(&nid)
		neighbouringCities = append(neighbouringCities, nid)
	}

	return
}

//ByMap Fetch cities tied to a map.
func ByMap(dbh *db.Handler, id int) (cities []*City, err error) {
	err = nil
	tmpCities := make(map[int]*City)
	rows := dbh.Query("select city_id from cities where map_id=$1", id)
	for rows.Next() {

		city := new(City)
		rows.Scan(&city.ID)

		tmpCities[city.ID] = city
	}

	for k, v := range tmpCities {
		// seek its neighbours
		rows = dbh.Query("select to_city_id from neighbouring_cities left outer join cities on from_city_id=city_id where city_id=$1", k)
		for rows.Next() {
			var nid int
			rows.Scan(&nid)
			v.Neighbours = append(v.Neighbours, tmpCities[nid])
		}

		cities = append(cities, v)
	}
	return
}

//Drop remove City from database
func (city *City) Drop(dbh *db.Handler) error {
	dbh.Query("delete from cities where city_id=$1", city.ID)
	dbh.Query("delete from neighbouring_cities where from_city_id=$1 or to_city_id=$2", city.ID, city.ID)
	log.Printf("City: Dropped City %d: %s from database", city.ID, "")
	city.ID = 0
	return nil
}
