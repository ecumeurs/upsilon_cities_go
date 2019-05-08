package city

import (
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
	Storage       storage.Storage
	LastUpdate    time.Time
	NextUpdate    time.Time

	RessourceProducers map[int]*producer.Producer
	ProductFactories   map[int]*producer.Producer

	ActiveRessourceProducers []*producer.Production
	ActiveProductFactories   []*producer.Production
}

//New create a new city ;)
func New() (city *City) {
	city = new(City)
	city.Storage = storage.New()
	city.NeighboursID = make([]int, 0)
	city.Roads = make([]node.Pathway, 0)
	city.RessourceProducers = make(map[int]*producer.Producer)
	city.ProductFactories = make(map[int]*producer.Producer)
	city.ActiveProductFactories = make([]*producer.Production, 0)
	city.ActiveRessourceProducers = make([]*producer.Production, 0)
	return
}

//CheckActivity Will check active producer for termination
//Will check inactive producers for activity start.
func (city *City) CheckActivity() error {
	futurNextUpdate := time.Now().UTC()
	nextUpdate := city.NextUpdate

	nActFact := make([]*producer.Production, 0)
	nActFactID := make(map[int]int)

	for _, v := range city.ActiveProductFactories {
		if v.IsFinished(nextUpdate) {
			producer.ProductionCompleted(&city.Storage, v, nextUpdate)
		} else {
			nActFact = append(nActFact, v)
			nActFactID[v.ProducerID] = v.ProducerID
			futurNextUpdate = tools.MinTime(futurNextUpdate, v.EndTime)
		}
	}
	nActRc := make([]*producer.Production, 0)
	nActRcID := make(map[int]int)
	for _, v := range city.ActiveRessourceProducers {
		if v.IsFinished(nextUpdate) {
			producer.ProductionCompleted(&city.Storage, v, nextUpdate)
		} else {
			nActRc = append(nActRc, v)
			nActRcID[v.ProducerID] = v.ProducerID
			futurNextUpdate = tools.MinTime(futurNextUpdate, v.EndTime)
		}
	}

	for k, v := range city.ProductFactories {
		// can start an already started factory ;)
		if _, found := nActFactID[k]; !found {
			prod, err := producer.Product(&city.Storage, v, nextUpdate)
			if err == nil {
				nActFact = append(nActFact, prod)
				nActFactID[k] = k
				futurNextUpdate = tools.MinTime(futurNextUpdate, prod.EndTime)
			}
		}
	}

	for k, v := range city.RessourceProducers {
		// can start an already started factory ;)
		if _, found := nActRcID[k]; !found {
			prod, err := producer.Product(&city.Storage, v, nextUpdate)
			if err == nil {
				nActRc = append(nActRc, prod)
				nActRcID[k] = k
				futurNextUpdate = tools.MinTime(futurNextUpdate, prod.EndTime)
			}
		}
	}

	city.ActiveProductFactories = nActFact
	city.ActiveRessourceProducers = nActRc
	city.LastUpdate = nextUpdate
	city.NextUpdate = futurNextUpdate

	if city.NextUpdate.Before(time.Now().UTC()) {
		return city.CheckActivity()
	}
	return nil
}
