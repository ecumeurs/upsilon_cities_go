package city

import (
	"fmt"
	"log"
	"time"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/cities/user_log"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/gameplay"
)

//StateHistory of city evolution
type StateHistory struct {
	Level        int
	IncreaseType int
	Message      string
	Date         time.Time
}

//State used by city_evolution
type State struct {
	CurrentLevel int

	MaxCaravans     int
	MaxRessources   int
	MaxFactories    int
	MaxResellers    int
	MaxStorageSpace int

	ProductionRate float32

	History []StateHistory

	Influence pattern.Pattern
}

//City stores city stuff
type City struct {
	ID              int
	Location        node.Point
	NeighboursID    []int
	CaravanID       []int
	Roads           []node.Pathway
	Name            string
	MapID           int
	CorporationID   int
	CorporationName string
	Storage         *storage.Storage
	LastUpdate      time.Time
	NextUpdate      time.Time

	HasStorageFull   bool
	StorageFullSince time.Time

	RessourceProducers map[int]*producer.Producer
	ProductFactories   map[int]*producer.Producer

	ActiveRessourceProducers map[int]*producer.Production
	ActiveProductFactories   map[int]*producer.Production

	CurrentMaxID int

	// Fame by CorporationID
	Fame map[int]int

	State State
}

//New create a new city ;)
func New() (city *City) {
	city = new(City)
	city.CorporationID = 0
	city.Storage = storage.New()
	city.HasStorageFull = false
	city.NeighboursID = make([]int, 0)
	city.Roads = make([]node.Pathway, 0)
	city.RessourceProducers = make(map[int]*producer.Producer)
	city.ProductFactories = make(map[int]*producer.Producer)
	city.ActiveProductFactories = make(map[int]*producer.Production, 0)
	city.ActiveRessourceProducers = make(map[int]*producer.Production, 0)
	city.Fame = make(map[int]int)
	log.Printf("City: Creating a new city !")

	city.State.History = make([]StateHistory, 0)
	city.State.Influence = pattern.Square

	city.NextUpdate = time.Now().UTC()
	city.LastUpdate = time.Now().UTC()
	return
}

//CheckActivity Will check active producer for termination
//Will check inactive producers for activity start.
func (city *City) CheckActivity(origin time.Time) (changed bool) {

	origin = tools.MaxTime(tools.RoundTime(origin), city.LastUpdate)

	if origin == city.LastUpdate {
		return false
	}

	// no need to check without an owner ...
	if city.CorporationID == 0 {
		return false
	}

	futurNextUpdate := origin.AddDate(0, 0, 1)
	nextUpdate := tools.MinTime(origin, city.NextUpdate)

	changed = false
	nActFact := make(map[int]*producer.Production)

	for _, v := range city.ActiveProductFactories {
		if v.IsFinished(nextUpdate) {
			producer.ProductionCompleted(city.Storage, v, nextUpdate)
			city.ProductFactories[v.ProducerID].Leveling(5)
			changed = true
		} else {
			nActFact[v.ProducerID] = v
			futurNextUpdate = tools.MinTime(futurNextUpdate, v.EndTime)
		}
	}
	nActRc := make(map[int]*producer.Production)
	for _, v := range city.ActiveRessourceProducers {
		if v.IsFinished(nextUpdate) {
			producer.ProductionCompleted(city.Storage, v, nextUpdate)
			city.RessourceProducers[v.ProducerID].Leveling(5)
			changed = true
		} else {
			nActRc[v.ProducerID] = v
			futurNextUpdate = tools.MinTime(futurNextUpdate, v.EndTime)
		}
	}

	fameLossBySpace := 0

	for k, v := range city.ProductFactories {
		// can start an already started factory ;)
		if _, found := nActFact[k]; !found {
			can, space, _ := producer.CanProduceShort(city.Storage, v)
			if can {
				prod, err := producer.Produce(city.Storage, v, nextUpdate)
				if err == nil {
					nActFact[k] = prod
					futurNextUpdate = tools.MinTime(futurNextUpdate, prod.EndTime)
					changed = true
				}
			} else {
				if space {
					// not enough space => fame loss

					fameLossBySpace++
				}
			}
		}
	}

	for k, v := range city.RessourceProducers {
		// can start an already started factory ;)
		if _, found := nActRc[k]; !found {
			can, space, _ := producer.CanProduceShort(city.Storage, v)
			if can {
				prod, err := producer.Produce(city.Storage, v, nextUpdate)
				if err == nil {
					nActRc[k] = prod
					futurNextUpdate = tools.MinTime(futurNextUpdate, prod.EndTime)
					changed = true
				}
			} else {
				if space {
					// not enough space => fame loss
					fameLossBySpace++
				}
			}
		}
	}

	if !city.HasStorageFull && fameLossBySpace > 0 {
		city.HasStorageFull = true
		city.StorageFullSince = nextUpdate

		city.AddFame(city.CorporationID, "can't produce", gameplay.GetInt("fame_loss_by_space", -3)*fameLossBySpace)
	}

	if fameLossBySpace > 0 {
		for nextUpdate.After(tools.AddCycles(city.StorageFullSince, 10)) {
			city.AddFame(city.CorporationID, "can't produce", gameplay.GetInt("fame_loss_by_space", -3)*fameLossBySpace)
			city.StorageFullSince = tools.AddCycles(city.StorageFullSince, 10)
		}
	} else {
		city.HasStorageFull = false
	}

	city.ActiveProductFactories = nActFact
	city.ActiveRessourceProducers = nActRc
	city.LastUpdate = nextUpdate
	city.NextUpdate = futurNextUpdate

	if city.Fame[city.CorporationID] < 50 {
		// critical city owner fame ... No need to continue producing ...
		return true
	}

	if city.NextUpdate.Before(origin) {
		return city.CheckActivity(origin) || changed
	}

	return
}

//AddFame update fame of city by provided margin.
func (city *City) AddFame(corpID int, message string, fameDiff int) {
	city.Fame[corpID] = city.Fame[corpID] + fameDiff
	if fameDiff >= 0 {
		user_log.NewFromCorp(corpID, user_log.UL_Info, fmt.Sprintf("City %s gain %d Fame (New: %d) for %s", city.Name, fameDiff, city.Fame[corpID], message))
	} else {
		user_log.NewFromCorp(corpID, user_log.UL_Warn, fmt.Sprintf("City %s loses %d Fame (New: %d) because %s", city.Name, -fameDiff, city.Fame[corpID], message))
		if city.Fame[corpID] < 100 {
			user_log.NewFromCorp(corpID, user_log.UL_Bad, fmt.Sprintf("City %s About to be lost", city.Name))
		}
	}
}

//CheckCityOwnership Checks city's owner fame, if fame drops below threshold of 50, owner is kicked. A check is then made to see if the owner can still continue play.
//returns true when everything is okay ;)
func (city *City) CheckCityOwnership(dbh *db.Handler) bool {
	if city.Fame[city.CorporationID] < 50 {
		log.Printf("City: %d %s Kick %d %s out", city.ID, city.Name, city.CorporationID, city.CorporationName)
		user_log.NewFromCorp(city.CorporationID, user_log.UL_Bad, fmt.Sprintf("City: %s Kick %s out", city.Name, city.CorporationName))

		city.CorporationID = 0
		city.CorporationName = "Uncorporated"
		city.Update(dbh)

		corp, err := corporation_manager.GetCorporationHandler(city.CorporationID)
		if err != nil {
			return false
		}

		corp.Call(func(corp *corporation.Corporation) {
			res := make([]int, 0)
			for _, v := range corp.CitiesID {
				if v == city.ID {

				} else {
					res = append(res, v)
				}
			}

			corp.CitiesID = res
		})

		if !corp.Get().IsViable() {
			log.Printf("Corporation: Lost its last city ...")
			return false
		}
		user_log.NewFromCorp(city.CorporationID, user_log.UL_Warn, fmt.Sprintf("Corporation still has %d cities", len(corp.Get().CitiesID)))
	}
	return true
}

//CanProduce tell whether city can produce item based on name.
func (city *City) CanProduce(itm item.Item) bool {
	for _, v := range city.RessourceProducers {
		for _, w := range v.Products {
			if w.ItemName == itm.Name {
				return true
			}
		}
	}
	for _, v := range city.ProductFactories {
		for _, w := range v.Products {
			if w.ItemName == itm.Name {
				return true
			}
		}
	}

	return false
}
