package producer

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
	"upsilon_cities_go/config"
	"upsilon_cities_go/lib/cities/tools"
)

//Factory describe a producer at level 0
type Factory struct {
	Quality      tools.IntRange
	Quantity     tools.IntRange
	BasePrice    int
	Requirements []requirement
	Delay        int // in cycles
	ItemType     []string
	ItemName     string
	IsRessource  bool `json:"-"`
	IsAdvanced   bool
	ProducerName string
	Origin       string `json:"-"`
}

// CreateSampleFile does what it says
func CreateSampleFile() {
	factories := make([]*Factory, 0)
	f := new(Factory)
	f.Quality.Min = 20
	f.Quality.Max = 25
	f.Quantity.Min = 1
	f.Quantity.Max = 2
	f.BasePrice = 6
	f.IsAdvanced = false
	f.Delay = 1

	f.ItemType = make([]string, 0)
	f.ItemType = append(f.ItemType, "ItemType3")
	f.ItemName = "TestItem"
	f.IsRessource = false
	f.ProducerName = "TestMaker"
	f.Requirements = make([]requirement, 0)
	var r requirement
	r.Ressource = "TestItemType2"
	r.Type = true
	r.Quantity = 3
	r.Quality.Min = 5
	r.Quality.Max = 50
	f.Requirements = append(f.Requirements, r)

	factories = append(factories, f)

	bytes, _ := json.MarshalIndent(factories, "", "\t")
	ioutil.WriteFile(fmt.Sprintf("%s/%s", config.DATA_PRODUCERS, "sample.json.sample"), bytes, 0644)
}

var knownProducers map[string][]*Factory
var knownProducersNames map[string][]*Factory
var ressources []string
var factories []string

//Load load factories
func Load() {
	knownProducers = make(map[string][]*Factory)
	knownProducersNames = make(map[string][]*Factory)
	ressources = make([]string, 0)
	factories = make([]string, 0)

	filepath.Walk(config.MakePath(config.DATA_PRODUCERS), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Producer: prevent panic by handling failure accessing a path %q: %v\n", config.DATA_PRODUCERS, err)
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

				p.Origin = info.Name()
				p.IsRessource = len(p.Requirements) == 0
				knownProducersNames[p.ItemName] = append(knownProducersNames[p.ItemName], p)

				for _, w := range p.ItemType {
					if p.IsRessource {
						ressources = append(ressources, w)
					} else {
						factories = append(factories, w)
					}

					knownProducers[w] = append(knownProducers[w], p)
				}

			}
		}

		return nil
	})

	validate()
	log.Printf("Producer: Loaded %d factories, %d ressource producers", len(factories), len(ressources))
}

//validate that all producers have valide requirements
func validate() {
	for _, v := range knownProducers {
		for _, vv := range v {
			log.Printf("Producer: Loaded: %s", vv.String())
			for _, req := range vv.Requirements {
				_, found := knownProducers[req.Ressource]
				if !found {
					_, found = knownProducersNames[req.Ressource]
					if !found {
						log.Printf("Producer: Invalid Producer registered: %s", vv.String())
						log.Fatalf("Producer: It misses required ressource: %s", req.String())
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

	state := ""
	if pf.IsAdvanced {
		state += "A"
	}
	if pf.IsRessource {
		state += "R"
	}
	return fmt.Sprintf("Factory: %s (%s) -> (%s:%s) %s", pf.ProducerName, strings.Join(reqs, ","), pf.ItemName, pf.ItemType, state)
}

func (pf *Factory) create() (prod *Producer) {
	prod = new(Producer)
	prod.ID = 0 // unset right now, will be the job of City to assign it an id.
	prod.ProductName = pf.ItemName
	prod.ProductType = pf.ItemType
	prod.Name = pf.ProducerName
	prod.Quality = pf.Quality
	prod.Quantity = pf.Quantity
	prod.Delay = pf.Delay
	prod.Requirements = pf.Requirements
	prod.BasePrice = pf.BasePrice
	prod.Advanced = pf.IsAdvanced
	prod.Level = 1
	prod.CurrentXP = 0
	prod.NextLevel = GetNextLevel(prod.Level)
	return prod
}

//CreateRandomBaseFactory pick from known base producer (not advanced).
func CreateRandomBaseFactory() *Producer {
	for true {
		rnd := rand.Intn(len(factories))
		rnd2 := rand.Intn(len(knownProducers[factories[rnd]]))
		fact := knownProducers[factories[rnd]][rnd2]
		if !fact.IsAdvanced {
			return fact.create()
		}
	}
	return nil
}

//CreateRandomRessource pick from known producer one.
func CreateRandomRessource() *Producer {
	rnd := rand.Intn(len(ressources))
	rnd2 := rand.Intn(len(knownProducers[ressources[rnd]]))
	return knownProducers[ressources[rnd]][rnd2].create()
}

//CreateRandomBaseRessource pick from known producer one.
func CreateRandomBaseRessource() *Producer {
	for true {
		rnd := rand.Intn(len(ressources))
		rnd2 := rand.Intn(len(knownProducers[ressources[rnd]]))
		fact := knownProducers[ressources[rnd]][rnd2]
		if !fact.IsAdvanced {
			return fact.create()
		}
	}
	return nil
}

//CreateProducer producer of matching type
func CreateProducer(item string) (*Producer, error) {
	if _, found := knownProducers[item]; !found {
		return nil, fmt.Errorf("item factory for %s unknown", item)
	}
	rnd2 := rand.Intn(len(knownProducers[item]))
	return knownProducers[item][rnd2].create(), nil
}

//CreateProducerByName producer of matching type
func CreateProducerByName(item string) (*Producer, error) {
	if _, found := knownProducersNames[item]; !found {
		return nil, fmt.Errorf("item factory for %s unknown", item)
	}
	rnd2 := rand.Intn(len(knownProducersNames[item]))
	return knownProducersNames[item][rnd2].create(), nil
}

//CreateFactoryNotAdvanced find a factory whose requirement contains at least one of items.
func CreateFactoryNotAdvanced(items map[string]bool) (*Producer, error) {

	log.Printf("Producer: Attempting to find a factory using %v", items)

	for _, v := range factories {
		for _, vv := range knownProducers[v] {
			if vv.IsAdvanced {
				continue
			}

			log.Printf("Producer: Checking %s", vv.String())
			for _, req := range vv.Requirements {
				if items[req.Ressource] {
					return vv.create(), nil
				}
			}
		}
	}

	return nil, errors.New("unable to find a factory matching requirements")
}
