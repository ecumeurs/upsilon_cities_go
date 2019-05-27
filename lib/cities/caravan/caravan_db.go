package caravan

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

type dbCaravan struct {
	Exported Object
	Imported Object

	LoadingDelay      int // in cycles
	TravelingDistance int // in nodes
	TravelingSpeed    int // in cycle => this default to 10

	Credits            int
	Store              *storage.Storage
	ExchangeRateLHS    int
	ExchangeRateRHS    int
	State              int
	SendQty            int
	ExportCompensation int // money sent with export to buy products.
	ImportCompensation int // money sent with export to buy products.

	Location node.Point

	Aborted       bool
	OriginDropped bool
	TargetDropped bool

	LastChange time.Time
	NextChange time.Time
	EndOfTerm  time.Time

	// also need to add storage stuff ... that are forcibly removed.
	CurrentMaxID int64
	Reservations map[int64]int
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
	tmp.State = caravan.State
	tmp.Location = caravan.Location
	tmp.Aborted = caravan.Aborted
	tmp.ExchangeRateLHS = caravan.ExchangeRateLHS
	tmp.ExchangeRateRHS = caravan.ExchangeRateRHS
	tmp.LastChange = caravan.LastChange
	tmp.NextChange = caravan.NextChange
	tmp.EndOfTerm = caravan.EndOfTerm
	tmp.CurrentMaxID = caravan.Store.CurrentMaxID
	tmp.Reservations = caravan.Store.Reservations
	tmp.OriginDropped = caravan.OriginDropped
	tmp.TargetDropped = caravan.TargetDropped

	tmp.SendQty = caravan.SendQty
	tmp.ExportCompensation = caravan.ExportCompensation
	tmp.ImportCompensation = caravan.ImportCompensation

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
	caravan.State = db.State
	caravan.Location = db.Location
	caravan.Aborted = db.Aborted
	caravan.LastChange = db.LastChange
	caravan.NextChange = db.NextChange
	caravan.EndOfTerm = db.EndOfTerm
	caravan.ExchangeRateLHS = db.ExchangeRateLHS
	caravan.ExchangeRateRHS = db.ExchangeRateRHS

	caravan.Store.CurrentMaxID = db.CurrentMaxID
	caravan.Store.Reservations = db.Reservations
	caravan.SendQty = db.SendQty
	caravan.ExportCompensation = db.ExportCompensation
	caravan.ImportCompensation = db.ImportCompensation
	caravan.OriginDropped = db.OriginDropped
	caravan.TargetDropped = db.TargetDropped

	return nil
}

//Reload a caravan from database
func (caravan *Caravan) Reload(dbh *db.Handler) {
	id := caravan.ID
	rows := dbh.Query(`select 
					   caravan_id, 
					   origin_corporation_id, originc.name as ocname, 
					   origin_city_id, originct.city_name as octname,  
					   target_corporation_id, targetc.name as tcname, 
					   target_city_id, targetct.city_name as tctname, 
					   state, c.map_id, c.data 
					   from caravans as c
					   left outer join corporations as originc on originc.corporation_id = origin_corporation_id
					   left outer join corporations as targetc on targetc.corporation_id = target_corporation_id
					   left outer join cities as originct on originct.city_id = origin_city_id
					   left outer join cities as targetct on targetct.city_id = target_city_id
					   where caravan_id=$1`, id)

	defer rows.Close()
	for rows.Next() {
		err := caravan.fill(rows)
		if err != nil {
			rows.Close()
		}
	}
}

//Insert a caravan in database
func (caravan *Caravan) Insert(dbh *db.Handler) error {

	if !caravan.IsValid() {
		return errors.New("can't insert invalid caravan into database")
	}
	if caravan.ID > 0 {
		return caravan.Update(dbh)
	}

	// ensure capacity of storage is appropriate
	caravan.Store.Capacity = tools.Max(caravan.Exported.Quantity.Max, caravan.Imported.Quantity.Max)

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
					   state, c.map_id, c.data 
					   from caravans  as c
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
					   state, c.map_id, c.data 
					   from caravans as c
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

//ByCityID a caravan from database
func ByCityID(dbh *db.Handler, id int) ([]*Caravan, error) {

	var caravans []*Caravan

	rows := dbh.Query(`select 
					   caravan_id, 
					   origin_corporation_id, originc.name as ocname, 
					   origin_city_id, originct.city_name as octname,  
					   target_corporation_id, targetc.name as tcname, 
					   target_city_id, targetct.city_name as tctname, 
					   state, c.map_id, c.data 
					   from caravans as c
					   left outer join corporations as originc on originc.corporation_id = origin_corporation_id
					   left outer join corporations as targetc on targetc.corporation_id = target_corporation_id
					   left outer join cities as originct on originct.city_id = origin_city_id
					   left outer join cities as targetct on targetct.city_id = target_city_id
					   where origin_city_id=$1 or target_city_id=$2`, id, id)

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
					   state, c.map_id, c.data 
					   from caravans  as c
					   left outer join corporations as originc on originc.corporation_id = origin_corporation_id
					   left outer join corporations as targetc on targetc.corporation_id = target_corporation_id
					   left outer join cities as originct on originct.city_id = origin_city_id
					   left outer join cities as targetct on targetct.city_id = target_city_id
					   where c.map_id=$1`, id)

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

//CorpDrop caravan from corp perspective.
func (caravan *Caravan) CorpDrop(dbh *db.Handler, corpID int) error {
	if caravan.CorpOriginID == corpID {
		caravan.OriginDropped = true
		caravan.Update(dbh)
		if caravan.OriginDropped && caravan.TargetDropped {
			corp, _ := corporation_manager.GetCorporationHandler(caravan.CorpOriginID)
			corp.Call(func(corp *corporation.Corporation) {
				res := make([]int, 0)
				for _, v := range corp.CaravanID {
					if v == caravan.ID {

					} else {
						res = append(res, v)
					}
				}

				corp.CaravanID = res
			})
			corp, _ = corporation_manager.GetCorporationHandler(caravan.CorpTargetID)
			corp.Call(func(corp *corporation.Corporation) {
				res := make([]int, 0)
				for _, v := range corp.CaravanID {
					if v == caravan.ID {

					} else {
						res = append(res, v)
					}
				}

				corp.CaravanID = res
			})
			caravan.Drop(dbh)
		}
		return nil
	}
	if caravan.CorpTargetID == corpID {
		caravan.TargetDropped = true
		caravan.Update(dbh)
		if caravan.OriginDropped && caravan.TargetDropped {

			corp, _ := corporation_manager.GetCorporationHandler(caravan.CorpOriginID)
			corp.Call(func(corp *corporation.Corporation) {
				res := make([]int, 0)
				for _, v := range corp.CaravanID {
					if v == caravan.ID {

					} else {
						res = append(res, v)
					}
				}

				corp.CaravanID = res
			})
			corp, _ = corporation_manager.GetCorporationHandler(caravan.CorpTargetID)
			corp.Call(func(corp *corporation.Corporation) {
				res := make([]int, 0)
				for _, v := range corp.CaravanID {
					if v == caravan.ID {

					} else {
						res = append(res, v)
					}
				}

				corp.CaravanID = res
			})

			caravan.Drop(dbh)
		}
		return nil
	}

	return errors.New("unable to drop as it doesn't belongs to corporation")
}
