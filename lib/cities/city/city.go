package city

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"upsilon_cities_go/config"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/cities/user_log"
	"upsilon_cities_go/lib/db"
)

//City
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

	baseFactory := producer.CreateRandomBaseFactory()
	baseFactory.ID = city.CurrentMaxID
	city.ProductFactories[city.CurrentMaxID] = baseFactory
	city.CurrentMaxID++

	nbRessources := 0
	log.Printf("City: Added base factory %d %s ! ", baseFactory.FactoryID, baseFactory.Name)

	ressourcesUsed := make(map[int]bool)
	ressourcesAvailable := make(map[string]bool)
	factoriesUsed := make(map[int]bool)

	factoriesUsed[baseFactory.FactoryID] = true

	for _, v := range baseFactory.Requirements {
		var baseRessource *producer.Producer
		var err error
		baseRessource, err = producer.CreateProducerByRequirement(v)

		if err != nil {
			log.Fatalf("Producer: Failed to generate ressource producer: %s: %s", v.Denomination, err)
		}
		baseRessource.ID = city.CurrentMaxID
		city.RessourceProducers[city.CurrentMaxID] = baseRessource
		city.CurrentMaxID++
		nbRessources++
		for _, w := range baseRessource.Products {
			for _, y := range w.ItemTypes {
				ressourcesAvailable[y] = true
			}
		}
		ressourcesUsed[baseRessource.FactoryID] = true
		log.Printf("City: Added base Ressource %s %+v ! ", baseRessource.Name, baseRessource.Products)

	}

	nbRessources = (rand.Intn(2) + 3) - nbRessources // number of ressources to generate still.
	for nbRessources > 0 {
		// ensure we don't get already used ressources ;)

		baseRessource := producer.CreateRandomBaseRessource()
		for ressourcesUsed[baseRessource.FactoryID] {
			baseRessource = producer.CreateRandomBaseRessource()
		}

		baseRessource.ID = city.CurrentMaxID
		city.RessourceProducers[city.CurrentMaxID] = baseRessource
		city.CurrentMaxID++
		nbRessources--
		for _, v := range baseRessource.Products {
			for _, w := range v.ItemTypes {
				ressourcesAvailable[w] = true
			}
		}
		ressourcesUsed[baseRessource.FactoryID] = true
		log.Printf("City: Completed with base Ressource %s %+v ! ", baseRessource.Name, baseRessource.Products)
	}

	nbFactories := rand.Intn(2) + 1

	if nbFactories == 2 {
		baseFactory := producer.CreateRandomBaseFactory()
		for factoriesUsed[baseFactory.FactoryID] {
			baseFactory = producer.CreateRandomBaseFactory()
		}

		factoriesUsed[baseFactory.FactoryID] = true
		baseFactory.ID = city.CurrentMaxID
		city.ProductFactories[city.CurrentMaxID] = baseFactory
		city.CurrentMaxID++
		nbFactories--

		log.Printf("City: Added with base Factory %d %s ! ", baseFactory.FactoryID, baseFactory.Name)
	}

	if nbFactories == 1 {
		baseFactory, err := producer.CreateFactoryNotAdvanced(ressourcesAvailable, factoriesUsed)

		if err != nil {
			log.Fatalf("City: Failed to build city due to %s", err)
		}

		baseFactory.ID = city.CurrentMaxID
		city.ProductFactories[city.CurrentMaxID] = baseFactory
		city.CurrentMaxID++
		nbFactories--
		log.Printf("City: Completed with base Factory %d %s ! ", baseFactory.FactoryID, baseFactory.Name)
	}

	city.NextUpdate = time.Now().UTC()
	city.LastUpdate = time.Now().UTC()
	city.CheckActivity(time.Now().UTC())
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
				prod, err := producer.Product(city.Storage, v, nextUpdate)
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
				prod, err := producer.Product(city.Storage, v, nextUpdate)
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

		city.AddFame(city.CorporationID, "can't produce", config.FAME_LOSS_BY_SPACE*fameLossBySpace)
		city.StorageFullSince = tools.AddCycles(city.StorageFullSince, 10)
	}

	if fameLossBySpace > 0 {
		for nextUpdate.After(tools.AddCycles(city.StorageFullSince, 10)) {
			city.AddFame(city.CorporationID, "can't produce", config.FAME_LOSS_BY_SPACE*fameLossBySpace)
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
	if fameDiff > 0 {
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
