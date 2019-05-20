package caravan

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/db"
)

type dbCaravan struct {
	Exported Object
	Imported Object

	LoadingDelay      int // in cycles
	TravelingDistance int // in nodes
	TravelingSpeed    int // in cycle => this default to 10

	Credits         int
	Store           *storage.Storage
	ExchangeRateLHS int
	ExchangeRateRHS int

	Location node.Point

	Aborted bool

	LastChange time.Time
	NextChange time.Time
	EndOfTerm  time.Time
}

func (caravan *Caravan) dbjsonify() (res []byte, err error) {
	err = nil
	var tmp dbCaravan

	tmp.Exported = caravan.Exported
	tmp.Imported = caravan.Imported
	tmp.LoadingDelay = caravan.LoadingDelay
	tmp.TravelingDistance = caravan.TravelingDistance
	tmp.TravelingSpeed = caravan.TravelingSpeed
	tmp.Credits = caravan.Credits
	tmp.Store = caravan.Store
	tmp.Location = caravan.Location
	tmp.Aborted = caravan.Aborted
	tmp.ExchangeRateLHS = caravan.ExchangeRateLHS
	tmp.ExchangeRateRHS = caravan.ExchangeRateRHS
	tmp.LastChange = caravan.LastChange
	tmp.NextChange = caravan.NextChange
	tmp.EndOfTerm = caravan.EndOfTerm

	return json.Marshal(tmp)
}

// reverse operation unpack json ;)
func (caravan *Caravan) dbunjsonify(fromJSON []byte) (err error) {
	var db dbCaravan
	err = json.Unmarshal(fromJSON, &db)
	if err != nil {
		return err
	}

	caravan.Exported = db.Exported
	caravan.Imported = db.Imported
	caravan.LoadingDelay = db.LoadingDelay
	caravan.TravelingDistance = db.TravelingDistance
	caravan.TravelingSpeed = db.TravelingSpeed
	caravan.Credits = db.Credits
	caravan.Store = db.Store
	caravan.Location = db.Location
	caravan.Aborted = db.Aborted
	caravan.LastChange = db.LastChange
	caravan.NextChange = db.NextChange
	caravan.EndOfTerm = db.EndOfTerm
	caravan.ExchangeRateLHS = db.ExchangeRateLHS
	caravan.ExchangeRateRHS = db.ExchangeRateRHS

	return nil
}

//Reload a caravan from database
func (caravan *Caravan) Reload(dbh *db.Handler) {
	caravan, _ = ByID(dbh, caravan.ID)
}

//Insert a caravan in database
func (caravan *Caravan) Insert(dbh *db.Handler) error {
	if !caravan.IsValid() {
		return errors.New("can't insert invalid caravan into database")
	}
	if caravan.ID > 0 {
		return caravan.Update(dbh)
	}

	rows := dbh.Query("insert into caravans(state, origin_corporation_id, target_corporation_id, origin_city_id, target_city_id, map_id) values(0, $1,$2,$3,$4,$5) returning caravan_id",
		caravan.CorpOriginID, caravan.CorpTargetID, caravan.CityOriginID, caravan.CityTargetID, caravan.MapID)

	for rows.Next() {
		rows.Scan(&caravan.ID)
	}
	rows.Close()

	log.Printf("Caravan: Inserted caravan into db.")
	return caravan.Update(dbh)
}

//Update a caravan in database
func (caravan *Caravan) Update(dbh *db.Handler) error {
	if !caravan.IsValid() {
		return errors.New("can't insert invalid caravan into database")
	}
	if caravan.ID == 0 {
		return caravan.Insert(dbh)
	}

	data, err := caravan.dbjsonify()
	if err != nil {
		return err
	}

	dbh.Query("update caravans set data=$1 where caravan_id=$2", data, caravan.ID).Close()

	return nil
}

//Drop a caravan from database
func (caravan *Caravan) Drop(dbh *db.Handler) error {
	return DropByID(dbh, caravan.ID)
}

//DropByID a caravan from database
func DropByID(dbh *db.Handler, id int) error {
	dbh.Query("delete from caravans where caravan_id=$1", id).Close()
	return nil
}

func (caravan *Caravan) fill(rows *sql.Rows) error {
	var data []byte

	rows.Scan(&caravan.ID,
		&caravan.CorpOriginID, &caravan.CorpOriginName,
		&caravan.CityOriginID, &caravan.CityOriginName,
		&caravan.CorpTargetID, &caravan.CorpTargetName,
		&caravan.CityTargetID, &caravan.CityTargetName,
		&caravan.State, &caravan.MapID, &data)

	return caravan.dbunjsonify(data)
}

//ByID a caravan from database
func ByID(dbh *db.Handler, id int) (*Caravan, error) {

	rows := dbh.Query(`select 
					   caravan_id, 
					   origin_corporation_id, originc.name as ocname, 
					   origin_city_id, originct.city_name as octname,  
					   target_corporation_id, targetc.name as tcname, 
					   target_city_id, targetct.city_name as tctname, 
					   state, map_id, data 
					   from caravans 
					   left outer join corporations as originc on originc.corporation_id = origin_corporation_id
					   left outer join corporations as targetc on targetc.corporation_id = target_corporation_id
					   left outer join cities as originct on originct.city_id = origin_city_id
					   left outer join cities as targetct on targetct.city_id = target_city_id
					   where caravan_id=$1`, id)

	defer rows.Close()
	for rows.Next() {
		caravan := new(Caravan)
		err := caravan.fill(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		return caravan, nil
	}
	return nil, fmt.Errorf("no caravan of id %d found", id)
}

//ByCorpID a caravan from database
func ByCorpID(dbh *db.Handler, id int) ([]*Caravan, error) {

	var caravans []*Caravan

	rows := dbh.Query(`select 
					   caravan_id, 
					   origin_corporation_id, originc.name as ocname, 
					   origin_city_id, originct.city_name as octname,  
					   target_corporation_id, targetc.name as tcname, 
					   target_city_id, targetct.city_name as tctname, 
					   state, map_id, data 
					   from caravans 
					   left outer join corporations as originc on originc.corporation_id = origin_corporation_id
					   left outer join corporations as targetc on targetc.corporation_id = target_corporation_id
					   left outer join cities as originct on originct.city_id = origin_city_id
					   left outer join cities as targetct on targetct.city_id = target_city_id
					   where origin_corporation_id=$1 or target_corporation_id=$2`, id, id)

	defer rows.Close()
	for rows.Next() {
		caravan := new(Caravan)
		err := caravan.fill(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		caravans = append(caravans, caravan)
	}

	if len(caravans) > 0 {
		return caravans, nil
	}
	return nil, fmt.Errorf("no caravan on map id %d found", id)
}

//ByMapID a caravan from database
func ByMapID(dbh *db.Handler, id int) ([]*Caravan, error) {

	var caravans []*Caravan

	rows := dbh.Query(`select 
					   caravan_id, 
					   origin_corporation_id, originc.name as ocname, 
					   origin_city_id, originct.city_name as octname,  
					   target_corporation_id, targetc.name as tcname, 
					   target_city_id, targetct.city_name as tctname, 
					   state, map_id, data 
					   from caravans 
					   left outer join corporations as originc on originc.corporation_id = origin_corporation_id
					   left outer join corporations as targetc on targetc.corporation_id = target_corporation_id
					   left outer join cities as originct on originct.city_id = origin_city_id
					   left outer join cities as targetct on targetct.city_id = target_city_id
					   where map_id=$1`, id)

	defer rows.Close()
	for rows.Next() {
		caravan := new(Caravan)
		err := caravan.fill(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		caravans = append(caravans, caravan)
	}

	if len(caravans) > 0 {
		return caravans, nil
	}
	return nil, fmt.Errorf("no caravan on map id %d found", id)
}

//IsValid tells whether caravan is fully completed or not.
func (caravan *Caravan) IsValid() bool {

	return caravan.CorpOriginID != 0 &&
		caravan.CorpTargetID != 0 &&
		caravan.MapID != 0 &&
		caravan.CityOriginID != 0 &&
		caravan.CityTargetID != 0
}
