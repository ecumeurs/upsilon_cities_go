package producer_generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/misc/config/system"
)

//Factory describe a producer at level 0
type Factory struct {
	Requirements []producer.Requirement
	Products     []producer.Product
	Delay        int  // in cycles
	IsRessource  bool `json:"-"`
	IsAdvanced   bool
	ProducerName string
	Origin       string `json:"-"`
	ID           int    `json:"-"`
}

// CreateSampleFile does what it says
func CreateSampleFile() {
	factories := make([]*Factory, 0)
	f := new(Factory)
	f.IsAdvanced = false
	f.Delay = 1
	f.IsRessource = false
	f.Requirements = make([]producer.Requirement, 0)
	f.ProducerName = "TestMaker"

	var p producer.Product
	p.ItemTypes = append(p.ItemTypes, "ItemType3")
	p.ItemName = "TestItem"
	p.Quality.Min = 20
	p.Quality.Max = 25
	p.Quantity.Min = 1
	p.Quantity.Max = 2
	f.Products = append(f.Products, p)

	var r producer.Requirement
	r.ItemTypes = []string{"TestItemType2"}
	r.Denomination = "Some item"
	r.Quantity = 3
	r.Quality.Min = 5
	r.Quality.Max = 50
	f.Requirements = append(f.Requirements, r)

	factories = append(factories, f)

	bytes, _ := json.MarshalIndent(factories, "", "\t")
	ioutil.WriteFile(fmt.Sprintf("%s/%s", system.Get("data_producers", "data/producers"), "sample.json.sample"), bytes, 0644)
}

// by type
var knownProducers map[string][]*Factory

// by name
var knownProducersNames map[string][]*Factory

// factories name.
var ressources []string

// factories name.
var factories []string

//Initialize environment
func Initialize() {
	knownProducers = make(map[string][]*Factory)
	knownProducersNames = make(map[string][]*Factory)
	ressources = make([]string, 0)
	factories = make([]string, 0)
}

//Load load factories
func Load() {
	Initialize()

	baseID := 0

	filepath.Walk(system.MakePath(system.Get("data_producers", "data/producers")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Producer: prevent panic by handling failure accessing a path %q: %v\n", system.Get("data_producers", "data/producers"), err)
			return err
		}
		if strings.HasSuffix(info.Name(), ".json") {
			f, ferr := os.Open(path)
			if ferr != nil {
				log.Fatalln("Producer: No Producer data file present")
			}

			producerJSON, ferr := ioutil.ReadAll(f)
			if ferr != nil {
				log.Fatalln("Producer: Data file found but unable to read it all.")
			}

			f.Close()

			prods := make([]*Factory, 0)
			json.Unmarshal(producerJSON, &prods)

			for _, p := range prods {
				baseID++

				loadFactory(p, baseID, info.Name())
				log.Printf("Producer loaded: %d %s", baseID, p.String())
			}
		}

		return nil
	})

	validate()
	log.Printf("Producer: Loaded %d factories, %d ressource producers", len(factories), len(ressources))
}

//loadFactory will store a factory in memory with appropriate links done.
func loadFactory(p *Factory, baseID int, origin string) {
	p.ID = baseID

	p.Origin = origin
	// a resource producer produce only one item, and doesn't require anything to produce.

	p.IsRessource = len(p.Requirements) == 0 && len(p.Products) == 1
	for _, v := range p.Products {
		knownProducersNames[v.ItemName] = append(knownProducersNames[v.ItemName], p)

		if p.IsRessource {
			if !tools.InStringList(v.ItemName, ressources) {
				ressources = append(ressources, v.ItemName)
			}
		} else {
			if !tools.InStringList(v.ItemName, factories) {
				factories = append(factories, v.ItemName)
			}
		}

		for _, w := range v.ItemTypes {

			knownProducers[w] = append(knownProducers[w], p)
		}
	}
}

//ProducerRequiringTypes tell which producers need requetested types
func ProducerRequiringTypes(types []string, exclusive bool) (req []*Factory) {
	if len(types) == 0 {
		return
	}

	for _, v := range factories {
		for _, f := range knownProducersNames[v] {
			onlyInAvailable := true
			atLeastOne := false
			for _, r := range f.Requirements {
				if tools.HasOneIn(r.ItemTypes, types) {
					atLeastOne = true
				} else {
					onlyInAvailable = false
				}
			}
			if atLeastOne {
				if exclusive {
					if onlyInAvailable {
						req = append(req, f)
					}
				} else {
					req = append(req, f)
				}
			} else if onlyInAvailable {
				req = append(req, f)
			}
		}
	}

	return
}

//ResourceProducerProducingTypes Seek out resource producer that will produce targeted resources only
func ResourceProducerProducingTypes(types []string) (req []*Factory) {
	for k := range knownProducers {
		for idx := range knownProducers[k] {
			if knownProducers[k][idx].IsRessource {
				for _, p := range knownProducers[k][idx].Products {
					if tools.StringListMatchOne(types, p.ItemTypes) {
						req = append(req, knownProducers[k][idx])
						break
					}
				}
			}
		}
	}
	return
}

//ProducerMatchingTypes tell what producers will produce these types. (among other)
func ProducerMatchingTypes(types []string) (req []*Factory, found bool) {
	// take the first types in the list, then check them all ...
	if len(types) == 0 {
		found = false
		return
	}

	tp := types[0]

	for _, v := range knownProducers[tp] {
		var ftypes []string
		for _, w := range v.Products {
			ftypes = append(ftypes, w.ItemTypes...)
		}

		if tools.ListInStringList(types, ftypes) {
			req = append(req, v)
		}
	}

	found = len(req) > 0
	return
}

func producerMatchingRequirement(rq producer.Requirement) (res []*Factory, found bool) {
	if len(rq.ItemTypes) > 0 {
		return ProducerMatchingTypes(rq.ItemTypes)
	}
	res, found = knownProducersNames[rq.ItemName]
	return
}

//validate that all producers have valide requirements
func validate() {
	for _, v := range knownProducers {
		for _, vv := range v {
			log.Printf("Producer: Loaded: %s", vv.String())
			for _, req := range vv.Requirements {
				ressourceFactories, found := ProducerMatchingTypes(req.ItemTypes)
				if !found {
					_, found = knownProducersNames[req.ItemName]
					if !found {
						log.Printf("Producer: Invalid Producer registered: %s", vv.String())
						log.Fatalf("Producer: It misses required ressource: %s", req.String())
					}
				}

				if !vv.IsAdvanced {
					// not advanced means ... requires only ressources to be built.

					oneRessource := false
					for _, rf := range ressourceFactories {
						if rf.IsRessource {
							oneRessource = true
							break
						}
					}

					if !oneRessource {
						log.Printf("Producer: Invalid Producer registered: %s", vv.String())
						log.Fatalf("Producer: Marked as not advanced but requires non ressources type: %s", req.String())
					}
				}
			}
		}

	}
}

func (pf *Factory) String() string {
	reqs := make([]string, 0)
	for _, v := range pf.Requirements {
		reqs = append(reqs, v.String())
	}
	prods := make([]string, 0)
	for _, v := range pf.Products {
		prods = append(prods, v.String())
	}

	state := ""
	if pf.IsAdvanced {
		state += "A"
	}
	if pf.IsRessource {
		state += "R"
	}

	return fmt.Sprintf("Factory: %s [%s] -> [%s] %s", pf.ProducerName, strings.Join(reqs, ","), strings.Join(prods, ","), state)
}

//Create generate a producer from a factory
func (pf *Factory) Create() (prod *producer.Producer) {
	prod = new(producer.Producer)
	prod.ID = 0 // unset right now, will be the job of City to assign it an id.
	prod.Products = make(map[int]producer.Product, 0)
	prod.History = make(map[int][]int, 0)
	for idx, v := range pf.Products {
		var p producer.Product
		p.ID = (idx + 1)
		p.ItemName = v.ItemName
		p.ItemTypes = v.ItemTypes
		p.Quality = v.Quality
		p.Quantity = v.Quantity
		p.BasePrice = v.BasePrice
		prod.Products[(idx + 1)] = p
	}
	prod.Name = pf.ProducerName
	prod.Delay = pf.Delay
	prod.Requirements = pf.Requirements
	prod.Advanced = pf.IsAdvanced
	prod.Level = 1
	prod.CurrentXP = 0
	prod.NextLevel = producer.GetNextLevel(prod.Level)
	prod.FactoryID = pf.ID
	prod.LastActivity = tools.RoundNow()
	return prod
}

//CreateRandomBaseFactory pick from known base producer (not advanced).
func CreateRandomBaseFactory() *producer.Producer {
	for true {
		rnd := rand.Intn(len(factories))
		rnd2 := rand.Intn(len(knownProducersNames[factories[rnd]]))
		fact := knownProducersNames[factories[rnd]][rnd2]
		if !fact.IsAdvanced {
			return fact.Create()
		}
	}
	return nil
}

//CreateRandomRessource pick from known producer one.
func CreateRandomRessource() *producer.Producer {
	for true {
		rnd := rand.Intn(len(ressources))
		rnd2 := rand.Intn(len(knownProducersNames[ressources[rnd]]))
		if knownProducersNames[ressources[rnd]][rnd2].IsRessource {
			return knownProducersNames[ressources[rnd]][rnd2].Create()
		}
	}
	return nil
}

//CreateRandomBaseRessource pick from known producer one.
func CreateRandomBaseRessource() *producer.Producer {
	for true {
		rnd := rand.Intn(len(ressources))
		rnd2 := rand.Intn(len(knownProducersNames[ressources[rnd]]))
		fact := knownProducersNames[ressources[rnd]][rnd2]
		if !fact.IsAdvanced && fact.IsRessource {
			return fact.Create()
		}
	}
	return nil
}

//CreateProducer producer of matching type
func CreateProducer(item string) (*producer.Producer, error) {
	if _, found := knownProducers[item]; !found {
		return nil, fmt.Errorf("item factory for %s unknown", item)
	}
	rnd2 := rand.Intn(len(knownProducers[item]))
	return knownProducers[item][rnd2].Create(), nil
}

//CreateProducerByRequirement producer of matching type
func CreateProducerByRequirement(rq producer.Requirement) (*producer.Producer, error) {
	prods, found := producerMatchingRequirement(rq)
	if !found {
		return nil, fmt.Errorf("No producers matching requirement %s", rq.String())
	}
	rnd2 := rand.Intn(len(prods))
	return prods[rnd2].Create(), nil
}

//CreateProducerByName producer of matching type
func CreateProducerByName(item string) (*producer.Producer, error) {
	if _, found := knownProducersNames[item]; !found {
		return nil, fmt.Errorf("item factory for %s unknown", item)
	}
	rnd2 := rand.Intn(len(knownProducersNames[item]))
	return knownProducersNames[item][rnd2].Create(), nil
}

//CreateFactoryNotAdvanced find a factory whose requirement contains at least one of items.
func CreateFactoryNotAdvanced(items map[string]bool, notin map[int]bool) (*producer.Producer, error) {

	log.Printf("Producer: Attempting to find a factory using %v", items)
	log.Printf("Producer: Attempting to find a factory not in %v", notin)

	for _, v := range factories {
		for _, vv := range knownProducersNames[v] {
			if vv.IsAdvanced {
				continue
			}
			if vv.IsRessource {
				continue
			}
			if notin[vv.ID] {
				continue
			}

			log.Printf("Producer: Checking %s", vv.String())
			for _, req := range vv.Requirements {
				foundOne := false
				for _, w := range req.ItemTypes {
					if items[w] {
						foundOne = true
						break
					}
				}
				if foundOne {
					return vv.Create(), nil
				}
			}
		}
	}

	return nil, errors.New("unable to find a factory matching requirements")
}
