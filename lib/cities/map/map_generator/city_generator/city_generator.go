package city_generator

import (
	"log"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city/producer_generator"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
)

const (
	excluded  = 0
	available = 1
)

//CityGenerator generate desert ahah
type CityGenerator struct {
	//Density number of cities in 10x10
	Density            tools.IntRange
	InfluenceRange     tools.IntRange
	ExploitedResources tools.IntRange
	FabricsRunning     tools.IntRange

	InitCaravans     int
	InitResellers    int
	InitStorageSpace int

	InitProductionRate float32
}

//Create a new desert generator with randomized conf
func Create() (mg CityGenerator) {
	mg.Density = tools.MakeIntRange(1, tools.RandInt(3, 5))
	mg.InfluenceRange = tools.MakeIntRange(1, 2)
	mg.ExploitedResources = tools.MakeIntRange(2, 3)
	mg.FabricsRunning = tools.MakeIntRange(1, 2)
	mg.InitCaravans = 3
	mg.InitResellers = 0
	mg.InitStorageSpace = 500
	return
}

//Level of the sub generator see Generator Level
func (mg CityGenerator) Level() map_level.GeneratorLevel {
	return map_level.Structure
}

func (mg CityGenerator) generateCity(gd *grid.CompoundedGrid, loc node.Point) {
	nd := gd.Get(loc)
	nd.Type = nodetype.CityNode
	gd.Set(nd)

	cty := city.New()
	cty.ID = len(gd.Base.Cities)
	gd.Base.Cities[cty.ID] = cty
	gd.Base.LocationToCity[loc.ToInt(gd.Base.Size)] = cty

	cty.Storage.SetSize(mg.InitStorageSpace)
	cty.State.MaxCaravans = mg.InitCaravans
	cty.State.MaxFactories = mg.FabricsRunning.Roll()
	cty.State.MaxRessources = mg.ExploitedResources.Roll()
	cty.State.MaxResellers = mg.InitResellers
	cty.State.MaxStorageSpace = mg.InitStorageSpace

	cty.State.Influence = pattern.GenerateAdjascentPattern(mg.InfluenceRange.Roll())

	// list exploitable resources

	activeResources := make([]string, 0)
	for _, v := range gd.SelectPattern(loc, cty.State.Influence) {
		for _, w := range v.Activated {
			activeResources = append(activeResources, w.Type)
		}
	}

	// select a fabric that may use any of the resource, this one will be forcibly added.
	// its resources as well.

	candidateFactories := producer_generator.ProducerRequiringTypes(activeResources, true)

	builtResources := 0
	builtFactories := 0

	if len(candidateFactories) > 0 {
		idx := tools.RandInt(0, len(candidateFactories)-1)
		factory := candidateFactories[idx]
		prod := factory.Create()

		reqResources := make(map[string]bool)
		for _, r := range factory.Requirements {
			resource, _ := tools.OneIn(r.ItemTypes, activeResources) // must have it, as we checked earlier
			if _, found := reqResources[resource]; !found {
				reqResources[resource] = true
			}
		}

		foundResources := make(map[string][]*producer_generator.Factory)
		foundMatchingResources := true

		for k := range reqResources {
			found := false
			res, _ := producer_generator.ProducerMatchingTypes([]string{k})
			for _, v := range res {
				if v.IsRessource {
					if _, known := foundResources[k]; !known {
						foundResources[k] = make([]*producer_generator.Factory, 0)
					}
					foundResources[k] = append(foundResources[k], v)
					found = true
				}
			}
			if !found {
				log.Printf("CG: Shouldn't happend, but map resources exist without a matching factory/producer of this resource. Type %v", k)
				foundMatchingResources = false
				break
			}
		}

		if !foundMatchingResources {
			log.Printf("CG: Can't generate factory as soem resources are missing ...")

		} else {
			// add factory of product
			cty.ProductFactories[cty.CurrentMaxID] = prod
			cty.CurrentMaxID++
			builtFactories = 1
			// add one of each resources.
			for _, v := range foundResources {
				idx := tools.RandInt(0, len(v))
				rprod := v[idx].Create()
				cty.RessourceProducers[cty.CurrentMaxID] = rprod
				cty.CurrentMaxID++
				builtResources++
			}
		}

	} else {
		log.Printf("CG: Weird got no candidates factories for resources: %v", activeResources)
	}

	for builtFactories < cty.State.MaxFactories {

	}

	for builtResources < cty.State.MaxRessources {

	}

	// for remaining resources, add at random
	// for remaining fabrics, add at random

	cty.CheckActivity(time.Now().UTC())

}

//Generate Will apply generator to provided grid
func (mg CityGenerator) Generate(gd *grid.CompoundedGrid) error {
	density := mg.Density.Roll()
	size := gd.Base.Size
	nb := (size / 10) * density

	log.Printf("CityGenerator: Attempting to add Cities to map density: %d Size: %d => number of cities to add: %d", density, size, nb)

	acc := gd.AccessibilityGrid()

	// no other cities within adj 3
	// city mustn't have any borders within adj 3 either
	// cities must be at least within 10 nodes of another.

	// note city placement and such may be regionalized and method of placement could differ
	// just as road would be.

	// here we will simply divide the map in sub square sectors, and within each sectors, we will select a city location.

	// First generate border exclusion pattern ;)

	borders := make(pattern.Pattern, 0, (gd.Base.Size*4)*2)
	for x := 0; x < gd.Base.Size; x++ {
		for y := 0; y < gd.Base.Size; y++ {
			if x <= 1 || x >= gd.Base.Size-2 {
				borders = append(borders, node.NP(x, y))
			} else if y <= 1 || y >= gd.Base.Size-2 {
				borders = append(borders, node.NP(x, y))
			}
		}
	}

	for _, v := range borders {
		if acc.IsAccessible(v) {
			acc.SetData(v, excluded)
		} else {
			acc.SetData(v, available)
		}
	}

	square := pattern.GenerateSquarePattern(2)
	refuse := pattern.GenerateAdjascentPattern(3)

	cities := make([]node.Point, 0)

	row := 5
	for row < gd.Base.Size {
		col := 5
		for col < gd.Base.Size {
			try := 0
			for try < 3 {
				idx := tools.RandInt(0, len(square))
				candidate := node.NP(col, row).Add(square[idx])
				if acc.GetData(candidate) == available {
					acc.Apply(candidate, refuse, func(n *node.Node, od int) (nd int) {
						nd = excluded
						return
					})
					mg.generateCity(gd, candidate)
					cities = append(cities, candidate)
				} else {
					try++
				}
			}
			col += 5
		}
		row += 5
	}

	return nil
}

//Name of the generator
func (mg CityGenerator) Name() string {
	return "CityGenerator"
}
