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
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
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
	Neighbours         tools.IntRange

	InitCaravans     int
	InitResellers    int
	InitStorageSpace int

	InitProductionRate float32
}

//Create a new desert generator with randomized conf
func Create() (mg CityGenerator) {
	mg.Density = tools.MakeIntRange(3, tools.RandInt(5, 7))
	mg.InfluenceRange = tools.MakeIntRange(1, 2)
	mg.ExploitedResources = tools.MakeIntRange(2, 3)
	mg.FabricsRunning = tools.MakeIntRange(1, 2)
	mg.Neighbours = tools.MakeIntRange(2, 4)
	mg.InitCaravans = 3
	mg.InitResellers = 0
	mg.InitStorageSpace = 500
	return
}

//Level of the sub generator see Generator Level
func (mg CityGenerator) Level() map_level.GeneratorLevel {
	return map_level.Structure
}

func (mg CityGenerator) generateCityPrepare(gd *grid.CompoundedGrid, dbh *db.Handler, loc node.Point) (cty *city.City) {
	gd.SetPCity(loc.X, loc.Y, true)

	cty = city.New()
	cty.Location = loc
	cty.MapID = gd.Base.ID
	cty.Insert(dbh)
	gd.Delta.Cities[cty.ID] = cty
	gd.Delta.LocationToCity[loc.ToInt(gd.Base.Size)] = cty

	cty.Storage.SetSize(mg.InitStorageSpace)
	cty.State.MaxCaravans = mg.InitCaravans
	cty.State.MaxFactories = mg.FabricsRunning.Roll()
	cty.State.MaxRessources = mg.ExploitedResources.Roll()
	cty.State.MaxResellers = mg.InitResellers
	cty.State.MaxStorageSpace = mg.InitStorageSpace

	cty.State.Influence = pattern.GenerateAdjascentPattern(mg.InfluenceRange.Roll())
	log.Printf("GC: Added city to %v", loc)

	return
}

func (mg CityGenerator) generateCity(gd *grid.CompoundedGrid, dbh *db.Handler, loc node.Point) {
	cty := mg.generateCityPrepare(gd, dbh, loc)
	// list exploitable resources

	log.Printf("%d cities in delta", len(gd.Delta.Cities))

	ar := make(map[string]bool, 0)
	for _, v := range gd.SelectPattern(loc, cty.State.Influence) {
		for _, w := range v.Activated {
			ar[w.Type] = true
		}
	}

	activeResources := make([]string, 0)
	for k := range ar {
		activeResources = append(activeResources, k)
	}

	// select a fabric that may use any of the resource, this one will be forcibly added.
	// its resources as well.

	log.Printf("GC: got active Resources: %v", activeResources)

	candidateResources := producer_generator.ResourceProducerProducingTypes(activeResources)
	log.Printf("GC: got candidate resources: %v", candidateResources)

	builtResources := 0

	buildResourcesTypes := make([]string, 0)

	if len(candidateResources) > 0 {
		for i := 0; i < cty.State.MaxRessources; i++ {
			idx := tools.RandInt(0, len(candidateResources)-1)
			fact := candidateResources[idx].Create()
			log.Printf("GC: Adding resource generator: %v %v", candidateResources[idx], fact)

			cty.RessourceProducers[cty.CurrentMaxID] = fact
			cty.CurrentMaxID++
			builtResources++
			buildResourcesTypes = append(buildResourcesTypes, candidateResources[idx].Products[0].ItemTypes...)
		}
	} else {
		log.Printf("CG: Weird got no candidates factories for resources: %v", activeResources)
	}

	// for remaining resources, add at random
	// for remaining fabrics, add at random

	candidateProducts := producer_generator.ProducerRequiringTypes(buildResourcesTypes, true)
	log.Printf("GC: got candidate products: %v", candidateProducts)

	if len(candidateProducts) > 0 {
		for i := 0; i < cty.State.MaxFactories; i++ {
			idx := tools.RandInt(0, len(candidateProducts)-1)
			fact := candidateProducts[idx].Create()
			log.Printf("GC: Adding product generator: %v %v", candidateProducts[idx], fact)

			cty.ProductFactories[cty.CurrentMaxID] = fact
			cty.CurrentMaxID++
		}
	} else {
		log.Printf("CG: Weird got no candidates factories for products: %v", candidateProducts)
	}

	cty.CheckActivity(time.Now().UTC())

}

//Generate Will apply generator to provided grid
func (mg CityGenerator) Generate(gd *grid.CompoundedGrid, dbh *db.Handler) error {
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

	for x := 0; x < gd.Base.Size; x++ {
		for y := 0; y < gd.Base.Size; y++ {
			pt := node.NP(x, y)
			if x <= 1 || x >= gd.Base.Size-2 {
				acc.SetData(pt, excluded)
			} else if y <= 1 || y >= gd.Base.Size-2 {
				acc.SetData(pt, excluded)
			} else {
				if acc.IsAccessible(pt) {
					acc.SetData(pt, available)
				} else {
					acc.SetData(pt, excluded)
				}
			}
		}
	}

	square := pattern.GenerateSquarePattern(2)
	refuse := pattern.GenerateAdjascentPattern(3)

	for retry := 0; retry < 3; retry++ {
		row := 5
		for row < gd.Base.Size {
			col := 5
			for col < gd.Base.Size {
				try := 0
				if tools.RandInt(0, 10) > 5 { // should build a city here ?
					for try < 3 {
						idx := tools.RandInt(0, len(square))
						candidate := node.NP(col, row).Add(square[idx])

						log.Printf("GC: Zone center: %v, candidate %v", node.NP(col, row), candidate)

						if acc.GetData(candidate) == available {
							acc.Apply(candidate, refuse, func(n *node.Node, od int) (nd int) {
								return excluded
							})
							mg.generateCity(gd, dbh, candidate)
							break
						} else {
							try++
						}
					}
				}
				if len(gd.Delta.Cities) == nb {
					break
				}
				col += 5
			}

			row += 5
			if len(gd.Delta.Cities) == nb {
				break
			}
		}
		if len(gd.Delta.Cities) == nb {
			break
		}
	}

	// find neighbours for each cities.
	for k, v := range gd.Delta.Cities {
		targetNeighbours := mg.Neighbours.Roll()

		distNgb := make(map[int]int)
		for kk, w := range gd.Delta.Cities {
			if kk == k {
				continue
			}
			distNgb[node.Distance(v.Location, w.Location)] = kk
		}

		testedNgb := make(map[int]bool)

		for _, w := range v.NeighboursID {
			testedNgb[w] = true
		}

		for _, w := range distNgb {
			if _, has := testedNgb[w]; !has {
				if len(gd.Delta.Cities[w].NeighboursID) < targetNeighbours {
					gd.Delta.Cities[w].NeighboursID = append(gd.Delta.Cities[w].NeighboursID, k)
					v.NeighboursID = append(v.NeighboursID, w)
					gd.Delta.Cities[w].Update(dbh)
				}
			}
			if len(v.NeighboursID) >= targetNeighbours {
				break
			}
		}

		v.Update(dbh)
	}

	return nil
}

//Name of the generator
func (mg CityGenerator) Name() string {
	return "CityGenerator"
}
