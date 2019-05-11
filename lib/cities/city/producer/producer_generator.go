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
	ItemType     string
	ItemName     string
	IsRessource  bool
	ProducerName string
}

// CreateSampleFile does what it says
func CreateSampleFile() {
	test := make(map[string][]*Factory)
	factories := make([]*Factory, 0)
	f := new(Factory)
	f.Quality.Min = 20
	f.Quality.Max = 25
	f.Quantity.Min = 1
	f.Quantity.Max = 2
	f.BasePrice = 6
	f.Delay = 1
	f.ItemType = "ItemType3"
	f.ItemName = "TestItem"
	f.IsRessource = false
	f.ProducerName = "TestMaker"
	f.Requirements = make([]requirement, 0)
	var r requirement
	r.RessourceType = "TestItemType2"
	r.Quantity = 3
	r.Quality.Min = 5
	r.Quality.Max = 50
	f.Requirements = append(f.Requirements, r)
	factories = append(factories, f)
	test["TestItemType"] = factories

	bytes, _ := json.MarshalIndent(test, "", "\t")
	ioutil.WriteFile(fmt.Sprintf("%s/%s", config.DATA_PRODUCERS, "sample.json.sample"), bytes, 0644)
}

var knownProducers map[string][]*Factory
var ressources []string
var factories []string

//Load load factories
func Load() {
	knownProducers = make(map[string][]*Factory)
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

			prods := make(map[string][]*Factory)
			json.Unmarshal(producerJSON, &prods)

			for k, v := range prods {
				base := knownProducers[k]

				first := base == nil || len(base) == 0

				if first {
					base = make([]*Factory, 0)
				}

				for _, p := range v {
					base = append(base, p)
				}

				if first {
					if base[0].IsRessource {
						ressources = append(ressources, k)
					} else {
						factories = append(factories, k)
					}
				}

				knownProducers[k] = base
			}
		}

		return nil
	})

	log.Printf("Producer: Loaded %d factories, %d ressource producers", len(factories), len(ressources))
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
	prod.Level = 1
	prod.CurrentXP = 0
	prod.NextLevel = 100
	return prod
}

//CreateRandomFactory pick from known producer one.
func CreateRandomFactory() *Producer {
	rnd := rand.Intn(len(factories))
	rnd2 := rand.Intn(len(knownProducers[factories[rnd]]))
	return knownProducers[factories[rnd]][rnd2].create()
}

//CreateRandomRessource pick from known producer one.
func CreateRandomRessource() *Producer {
	rnd := rand.Intn(len(ressources))
	rnd2 := rand.Intn(len(knownProducers[ressources[rnd]]))
	return knownProducers[ressources[rnd]][rnd2].create()
}

//CreateProducer producer of matching type
func CreateProducer(item string) (*Producer, error) {
	if _, found := knownProducers[item]; !found {
		return nil, fmt.Errorf("item factory for %s unknown", item)
	}
	rnd2 := rand.Intn(len(knownProducers[item]))
	return knownProducers[item][rnd2].create(), nil
}

//CreateFactory find a factory whose requirement contains at least one of items.
func CreateFactory(items map[string]string) (*Producer, error) {
	for _, v := range factories {
		for _, vv := range knownProducers[v] {
			for _, req := range vv.Requirements {
				if _, found := items[req.RessourceType]; found {
					return vv.create(), nil
				}
			}
		}
	}

	return nil, errors.New("unable to find a factory matching requirements")
}
