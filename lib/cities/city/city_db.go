package city

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/db"

	"github.com/lib/pq"
)

//Insert Stores or Update city as necessary. Won't insert neighbours at that time ;)
func (city *City) Insert(dbh *db.Handler) error {
	if city.ID <= 0 {
		city.LastUpdate = time.Now().UTC()
		city.NextUpdate = time.Now().UTC()
		// this is a new city. Simply attribute it a new ID
		res, err := dbh.Query("insert into cities(city_name, map_id, updated_at) values($1, $2, $3) returning city_id;", city.Name, city.MapID, city.LastUpdate)
		if err != nil {
			return fmt.Errorf("City DB : Insert Stores or Update city : %s", err)
		}
		for res.Next() {
			res.Scan(&city.ID)
		}

		res.Close()

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
	}

	// prepare json dump
	json, err := city.dbjsonify()
	if err != nil {
		return err
	}

	var res *sql.Rows
	if city.CorporationID == 0 {
		// dhb.Query has formater stuff ;)
		res, err = dbh.Query(`update cities set 
			city_name=$1
			, data=$2
			, corporation_id=NULL
			, updated_at=$3
			where city_id=$4;`, city.Name, json, city.LastUpdate, city.ID)
	} else {
		// dhb.Query has formater stuff ;)
		res, err = dbh.Query(`update cities set 
			city_name=$1
			, data=$2
			, corporation_id=$3
			, updated_at=$4
			where city_id=$5;`, city.Name, json, city.CorporationID, city.LastUpdate, city.ID)
	}
	if err != nil {
		err = fmt.Errorf("City DB : Failed to Update City : %s", err)
		return err
	}
	for res.Next() {
		res.Scan(&city.ID)
	}
	res.Close()

	err = city.dbCheckNeighbours(dbh)
	if err != nil {
		return err
	}

	log.Printf("City: Updated City %d: %s to database", city.ID, "")

	return nil
}

func (city *City) dbCheckNeighbours(dbh *db.Handler) error {
	// select known neighbours
	neighboursRows, err := dbh.Query("select to_city_id from neighbouring_cities where from_city_id=$1", city.ID)

	if err != nil {
		err = fmt.Errorf("City DB : Failed to select known neighbours (dbCheckNeighbours) : %s", err)
		return err
	}

	neighbours := make(map[int]bool)
	for neighboursRows.Next() {
		var id int
		neighboursRows.Scan(&id)
		neighbours[id] = true
	}
	neighboursRows.Close()
	// compare with actual neighbours

	var newNeighbours []int
	var missingNeighbours []int
	oldies := make(map[int]bool)
	for _, v := range city.NeighboursID {
		if !neighbours[v] {
			// it's not in database, but it's in city, it's new.
			newNeighbours = append(newNeighbours, v)
		}

		oldies[v] = true
	}

	for k := range neighbours {
		if !oldies[k] {
			missingNeighbours = append(missingNeighbours, k)
		}
	}

	if len(missingNeighbours) > 0 {
		// drop disappeared neighbours
		query, err := dbh.Query("delete from neighbouring_cities where from_city_id=$1 and to_city_id=ANY($2)", city.ID, pq.Array(missingNeighbours))
		if err != nil {
			err = fmt.Errorf("City DB : Failed to drop disappeared neighbours (dbCheckNeighbours) : %s", err)
			return err
		}
		query.Close()
		// be nice and remove reverse as well ;)
		query, _ = dbh.Query("delete from neighbouring_cities where to_city_id=$1 and from_city_id=ANY($2)", city.ID, pq.Array(missingNeighbours))
		if err != nil {
			err = fmt.Errorf("City DB : Failed to drop reverse disappeared neighbours (dbCheckNeighbours) : %s", err)
			return err
		}
		query.Close()
	}

	// add missing neighbours

	if len(newNeighbours) > 0 {
		log.Printf("City: About to insert %d / %d neighbours for city: %d", len(newNeighbours), len(city.NeighboursID), city.ID)
		for _, v := range newNeighbours {
			query, err := dbh.Query("insert into neighbouring_cities(to_city_id, from_city_id) values ($1,$2)", v, city.ID)
			if err != nil {
				err = fmt.Errorf("City DB : Failed to insert missing neighbours (dbCheckNeighbours) : %s", err)
				return err
			}
			query.Close()
		}
	}

	return nil
}

type dbCity struct {
	Location            node.Point
	Storage             *storage.Storage
	Roads               []node.Pathway
	FactoryCurrentMaxID int

	// also need to add storage stuff ... that are forcibly removed.
	CurrentMaxID int64
	Reservations map[int64]int

	LastUpdate time.Time
	NextUpdate time.Time

	RessourceProducers map[int]*producer.Producer
	ProductFactories   map[int]*producer.Producer

	ActiveRessourceProducers map[int]*producer.Production
	ActiveProductFactories   map[int]*producer.Production

	HasStorageFull   bool
	StorageFullSince time.Time

	Fame map[int]int
}

// prepare the json version for database, may not be the appropriate one for API ;)
// Useless for now, but will be usefull when city get storage and stuff like that.
func (city *City) dbjsonify() (res []byte, err error) {
	err = nil
	var tmp dbCity
	tmp.Location = city.Location
	tmp.Storage = city.Storage
	tmp.Roads = city.Roads
	tmp.FactoryCurrentMaxID = city.CurrentMaxID
	tmp.RessourceProducers = city.RessourceProducers
	tmp.ProductFactories = city.ProductFactories
	tmp.ActiveProductFactories = city.ActiveProductFactories
	tmp.ActiveRessourceProducers = city.ActiveRessourceProducers
	tmp.NextUpdate = city.NextUpdate
	tmp.CurrentMaxID = city.Storage.CurrentMaxID
	tmp.Reservations = city.Storage.Reservations
	tmp.Fame = city.Fame
	tmp.HasStorageFull = city.HasStorageFull
	tmp.StorageFullSince = city.StorageFullSince

	return json.Marshal(tmp)
}

// reverse operation unpack json ;)
func (city *City) dbunjsonify(fromJSON []byte) (err error) {
	var db dbCity
	err = json.Unmarshal(fromJSON, &db)
	if err != nil {
		return err
	}

	city.Location = db.Location
	city.Storage = db.Storage
	city.RessourceProducers = db.RessourceProducers
	city.ProductFactories = db.ProductFactories
	city.ActiveProductFactories = db.ActiveProductFactories
	city.ActiveRessourceProducers = db.ActiveRessourceProducers
	city.NextUpdate = db.NextUpdate

	city.Storage.CurrentMaxID = db.CurrentMaxID
	city.Storage.Reservations = db.Reservations
	city.Roads = db.Roads
	city.CurrentMaxID = db.FactoryCurrentMaxID
	city.Fame = db.Fame
	city.HasStorageFull = db.HasStorageFull
	city.StorageFullSince = db.StorageFullSince

	return nil
}

//Reload city ;)
func (city *City) Reload(dbh *db.Handler) {
	id := city.ID
	rows, err := dbh.Query("select city_id, c.map_id, c.data, updated_at, city_name, corporation_id, corp.name from cities as c left outer join corporations as corp using(corporation_id) where c.city_id=$1", id)

	if err != nil {
		log.Fatalf("City DB : Failed to select City for reload : %s ", err)
	}

	for rows.Next() {
		// hopefully there is only one ;) city_id is supposed to be unique.
		// atm only read city_id ;)
		var data []byte
		rows.Scan(&city.ID, &city.MapID, &data, &city.LastUpdate, &city.Name, &city.CorporationID, &city.CorporationName)
		city.dbunjsonify(data)
	}

	rows.Close()

	// seek its neighbours
	rows, err = dbh.Query("select to_city_id from neighbouring_cities where from_city_id=$1", id)

	if err != nil {
		log.Fatalf("City DB : Failed to select neighbouring_cities for reload : %s ", err)
	}

	for rows.Next() {
		var nid int
		rows.Scan(&nid)
		city.NeighboursID = append(city.NeighboursID, nid)
	}

	rows.Close()
	// seek its neighbours
	rows, err = dbh.Query("select caravan_id from caravans where origin_city_id=$1 or target_city_id=$2", id, id)

	if err != nil {
		log.Fatalf("City DB : Failed to select caravans for reload : %s ", err)
	}

	for rows.Next() {
		var nid int
		rows.Scan(&nid)
		city.CaravanID = append(city.CaravanID, nid)
	}

	rows.Close()
}

//ByID Fetch a city by id; note, won't load neighbouring cities ... or maybe only their ids ? ...
func ByID(dbh *db.Handler, id int) (city *City, err error) {

	err = nil
	city = new(City)

	rows, err := dbh.Query("select city_id, c.map_id, c.data, updated_at, city_name, corporation_id, corp.name from cities as c left outer join corporations as corp using(corporation_id) where city_id=$1", id)

	if err != nil {
		return nil, fmt.Errorf("City DB : Failed to get select city (ById) : %s", err)
	}

	for rows.Next() {
		// hopefully there is only one ;) city_id is supposed to be unique.
		// atm only read city_id ;)
		var data []byte
		rows.Scan(&city.ID, &city.MapID, &data, &city.LastUpdate, &city.Name, &city.CorporationID, &city.CorporationName)
		city.dbunjsonify(data)
	}

	rows.Close()

	// seek its neighbours
	rows, err = dbh.Query("select to_city_id from neighbouring_cities where from_city_id=$1", city.ID)

	if err != nil {
		return city, fmt.Errorf("City DB : Failed to get select neighbouring_cities (ById) : %s", err)
	}

	for rows.Next() {
		var nid int
		rows.Scan(&nid)
		city.NeighboursID = append(city.NeighboursID, nid)
	}

	rows.Close()
	// seek related caravan
	rows, err = dbh.Query("select caravan_id from caravans where origin_city_id=$1 or target_city_id=$2", city.ID, city.ID)

	if err != nil {
		return city, fmt.Errorf("City DB : Failed to get caravan (ByID) : %s", err)
	}

	for rows.Next() {
		var nid int
		rows.Scan(&nid)
		city.CaravanID = append(city.CaravanID, nid)
	}

	rows.Close()

	return
}

//ByMap Fetch cities tied to a map.
func ByMap(dbh *db.Handler, id int) (cities map[int]*City, err error) {
	err = nil
	cities = make(map[int]*City)

	rows, err := dbh.Query("select city_id, c.map_id, c.data, updated_at, city_name, corporation_id , corp.name from cities as c left outer join corporations as corp using(corporation_id) where c.map_id=$1", id)

	if err != nil {
		return nil, fmt.Errorf("City DB : Failed to get select city (ByMap) : %s", err)
	}

	for rows.Next() {

		city := new(City)
		var data []byte
		rows.Scan(&city.ID, &city.MapID, &data, &city.LastUpdate, &city.Name, &city.CorporationID, &city.CorporationName)
		city.dbunjsonify(data)

		cities[city.ID] = city
	}

	rows.Close()

	log.Printf("City: Found %d cities in map %d", len(cities), id)

	for k, v := range cities {
		// seek its neighbours
		rows, err = dbh.Query("select to_city_id from neighbouring_cities left outer join cities on from_city_id=city_id where city_id=$1", k)

		if err != nil {
			return cities, fmt.Errorf("City DB : Failed to get select neighbouring_cities (ByMap) : %s", err)
		}

		for rows.Next() {
			var nid int
			rows.Scan(&nid)
			v.NeighboursID = append(v.NeighboursID, nid)

		}

		rows.Close()

		// seek related caravan
		rows, err = dbh.Query("select caravan_id from caravans where origin_city_id=$1 or target_city_id=$2", k, k)

		if err != nil {
			return cities, fmt.Errorf("City DB : Failed to get caravan (ByMap) : %s", err)
		}

		for rows.Next() {
			var nid int
			rows.Scan(&nid)
			v.CaravanID = append(v.CaravanID, nid)
		}

		rows.Close()

		cities[k] = v
	}
	return
}

//Drop remove City from database
func (city *City) Drop(dbh *db.Handler) (err error) {

	_, err = dbh.Query("delete from cities where city_id=$1", city.ID)
	if err != nil {
		return fmt.Errorf("City DB : Failed to Drop City : %s", err)
	}

	_, err = dbh.Query("delete from neighbouring_cities where from_city_id=$1 or to_city_id=$2", city.ID, city.ID)
	if err != nil {
		return fmt.Errorf("City DB : Failed to Drop City neighbouring_cities : %s", err)
	}

	log.Printf("City: Dropped City %d: %s from database", city.ID, "")
	city.ID = 0
	return nil
}
