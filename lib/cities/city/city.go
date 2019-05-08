package city

import (
	"log"
	"math/rand"
	"time"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
)

//City
type City struct {
	ID            int
	Location      node.Point
	NeighboursID  []int
	Roads         []node.Pathway
	Name          string
	CorporationID int
	Storage       *storage.Storage
	LastUpdate    time.Time
	NextUpdate    time.Time

	RessourceProducers map[int]*producer.Producer
	ProductFactories   map[int]*producer.Producer

	ActiveRessourceProducers map[int]*producer.Production
	ActiveProductFactories   map[int]*producer.Production

	CurrentMaxID int
}

//New create a new city ;)
func New() (city *City) {
	city = new(City)
	city.Storage = storage.New()
	city.NeighboursID = make([]int, 0)
	city.Roads = make([]node.Pathway, 0)
	city.RessourceProducers = make(map[int]*producer.Producer)
	city.ProductFactories = make(map[int]*producer.Producer)
	city.ActiveProductFactories = make(map[int]*producer.Production, 0)
	city.ActiveRessourceProducers = make(map[int]*producer.Production, 0)

	baseFactory := producer.CreateRandomFactory()
	baseFactory.ID = city.CurrentMaxID
	city.ProductFactories[city.CurrentMaxID] = baseFactory
	city.CurrentMaxID++

	nbRessources := 0

	ressourcesAvailable := make(map[string]string)

	for _, v := range baseFactory.Requirements {
		baseRessource, err := producer.CreateProducer(v.RessourceType)
		if err != nil {
			log.Fatalf("Producer: Failed to generate ressource producer: %s: %s", v.RessourceType, err)
		}
		baseRessource.ID = city.CurrentMaxID
		city.RessourceProducers[city.CurrentMaxID] = baseRessource
		city.CurrentMaxID++
		nbRessources++
		ressourcesAvailable[baseRessource.ProductType] = baseRessource.ProductType
	}

	nbRessources = (rand.Intn(2) + 3) - nbRessources // number of ressources to generate still.
	for nbRessources > 0 {
		baseRessource := producer.CreateRandomRessource()
		baseRessource.ID = city.CurrentMaxID
		city.RessourceProducers[city.CurrentMaxID] = baseRessource
		city.CurrentMaxID++
		nbRessources--
		ressourcesAvailable[baseRessource.ProductType] = baseRessource.ProductType
	}

	nbFactories := rand.Intn(2) + 1

	if nbFactories == 2 {
		baseFactory := producer.CreateRandomFactory()
		baseFactory.ID = city.CurrentMaxID
		city.ProductFactories[city.CurrentMaxID] = baseFactory
		city.CurrentMaxID++
		nbFactories--
	}
	if nbFactories == 1 {
		baseFactory, _ := producer.CreateFactory(ressourcesAvailable)
		baseFactory.ID = city.CurrentMaxID
		city.ProductFactories[city.CurrentMaxID] = baseFactory
		city.CurrentMaxID++
		nbFactories--
	}

	city.NextUpdate = time.Now().UTC()
	city.LastUpdate = time.Now().UTC()
	city.CheckActivity(time.Now().UTC())
	return
}

//CheckActivity Will check active producer for termination
//Will check inactive producers for activity start.
func (city *City) CheckActivity(origin time.Time) (changed bool) {
	futurNextUpdate := origin.AddDate(0, 0, 1)
	nextUpdate := tools.MinTime(origin, city.NextUpdate)

	changed = false
	nActFact := make(map[int]*producer.Production)

	for _, v := range city.ActiveProductFactories {
		if v.IsFinished(nextUpdate) {
			producer.ProductionCompleted(city.Storage, v, nextUpdate)
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
			changed = true
		} else {
			nActRc[v.ProducerID] = v
			futurNextUpdate = tools.MinTime(futurNextUpdate, v.EndTime)
		}
	}

	for k, v := range city.ProductFactories {
		// can start an already started factory ;)
		if _, found := nActFact[k]; !found {
			prod, err := producer.Product(city.Storage, v, nextUpdate)
			if err == nil {
				nActFact[k] = prod
				futurNextUpdate = tools.MinTime(futurNextUpdate, prod.EndTime)
				changed = true
			}
		}
	}

	for k, v := range city.RessourceProducers {
		// can start an already started factory ;)
		if _, found := nActRc[k]; !found {
			prod, err := producer.Product(city.Storage, v, nextUpdate)
			if err == nil {
				nActRc[k] = prod
				futurNextUpdate = tools.MinTime(futurNextUpdate, prod.EndTime)
				changed = true
			}
		}
	}

	city.ActiveProductFactories = nActFact
	city.ActiveRessourceProducers = nActRc
	city.LastUpdate = nextUpdate
	city.NextUpdate = futurNextUpdate

	if city.NextUpdate.Before(origin) {
		return city.CheckActivity(origin) || changed
	}

	return
}
