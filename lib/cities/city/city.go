package city

import (
	"log"
	"math/rand"
	"time"
	"upsilon_cities_go/config"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

//City
type City struct {
	ID              int
	Location        node.Point
	NeighboursID    []int
	Roads           []node.Pathway
	Name            string
	CorporationID   int
	CorporationName string
	Storage         *storage.Storage
	LastUpdate      time.Time
	NextUpdate      time.Time

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
	city.NeighboursID = make([]int, 0)
	city.Roads = make([]node.Pathway, 0)
	city.RessourceProducers = make(map[int]*producer.Producer)
	city.ProductFactories = make(map[int]*producer.Producer)
	city.ActiveProductFactories = make(map[int]*producer.Production, 0)
	city.ActiveRessourceProducers = make(map[int]*producer.Production, 0)
	city.Fame = make(map[int]int)

	baseFactory := producer.CreateRandomBaseFactory()
	baseFactory.ID = city.CurrentMaxID
	city.ProductFactories[city.CurrentMaxID] = baseFactory
	city.CurrentMaxID++

	nbRessources := 0

	ressourcesAvailable := make(map[string]bool)
	factoriesAvailable := make(map[string]bool)

	for _, v := range baseFactory.Products {
		for _, w := range v.ItemTypes {
			factoriesAvailable[w] = true
		}
	}

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

	}

	nbRessources = (rand.Intn(2) + 3) - nbRessources // number of ressources to generate still.
	for nbRessources > 0 {
		baseRessource := producer.CreateRandomBaseRessource()
		// ensure we don't get already used ressources ;)
		var types []string
		for _, v := range baseRessource.Products {
			types = append(types, v.ItemTypes...)
		}

		for tools.ListInStringMap(types, ressourcesAvailable, true) {
			baseRessource = producer.CreateRandomBaseRessource()
			types = nil
			for _, v := range baseRessource.Products {
				types = append(types, v.ItemTypes...)
			}
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
	}

	nbFactories := rand.Intn(2) + 1

	if nbFactories == 2 {
		baseFactory := producer.CreateRandomBaseFactory()
		var types []string
		for _, v := range baseFactory.Products {
			types = append(types, v.ItemTypes...)
		}

		// ensure we don't get already used factories ;)
		for tools.ListInStringMap(types, factoriesAvailable, false) {
			baseFactory = producer.CreateRandomBaseFactory()
			types = nil
			for _, v := range baseFactory.Products {
				types = append(types, v.ItemTypes...)
			}

		}
		baseFactory.ID = city.CurrentMaxID
		city.ProductFactories[city.CurrentMaxID] = baseFactory
		city.CurrentMaxID++
		nbFactories--
		for _, v := range baseFactory.Products {
			for _, w := range v.ItemTypes {
				factoriesAvailable[w] = true
			}
		}
	}

	if nbFactories == 1 {
		baseFactory, _ := producer.CreateFactoryNotAdvanced(ressourcesAvailable)

		log.Printf("Test %+v ", baseFactory)
		var types []string
		for _, v := range baseFactory.Products {
			types = append(types, v.ItemTypes...)
		}

		// ensure we don't get already used factories ;)
		for tools.ListInStringMap(types, factoriesAvailable, true) {
			baseFactory = producer.CreateRandomBaseFactory()
			types = nil
			for _, v := range baseFactory.Products {
				types = append(types, v.ItemTypes...)
			}

		}
		baseFactory.ID = city.CurrentMaxID
		city.ProductFactories[city.CurrentMaxID] = baseFactory
		city.CurrentMaxID++
		nbFactories--
		for _, v := range baseFactory.Products {
			for _, w := range v.ItemTypes {
				factoriesAvailable[w] = true
			}
		}
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

	fameLossBySpace := false

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

					fameLossBySpace = true
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
					fameLossBySpace = true
				}
			}
		}
	}

	if fameLossBySpace {
		city.AddFame(city.CorporationID, config.FAME_LOSS_BY_SPACE*tools.CyclesBetween(nextUpdate, futurNextUpdate))
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
func (city *City) AddFame(corpID, fameDiff int) {
	city.Fame[corpID] = city.Fame[corpID] + fameDiff
}

//CheckCityOwnership Checks city's owner fame, if fame drops below threshold of 50, owner is kicked. A check is then made to see if the owner can still continue play.
//returns true when everything is okay ;)
func (city *City) CheckCityOwnership(dbh *db.Handler) bool {
	if city.Fame[city.CorporationID] < 50 {
		log.Printf("City: %d %s Kick %d %s out", city.ID, city.Name, city.CorporationID, city.CorporationName)
		city.CorporationID = 0
		city.Update(dbh)

		corp, err := corporation.ByID(dbh, city.CorporationID)
		if err != nil {
			return false
		}

		if !corp.IsViable() {
			log.Printf("Corporation: Lost its last city ...")
			corp.Drop(dbh)
			return false
		}
	}
	return true
}
