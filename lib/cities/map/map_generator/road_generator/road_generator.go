package road_generator

import (
	"log"
	"math/rand"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

const (
	north int = 0
	east  int = 1
	south int = 2
	west  int = 3
)

//RoadGenerator generate roads ahah
type RoadGenerator struct {
	RoadInfluenceCost int

	DefaultDepthReach int
	DefaultDepthCost  int

	Connection      map[int]bool
	CostDepth       map[nodetype.GroundType]int
	CostReach       map[nodetype.LandscapeType]int
	LTCostDepth     map[nodetype.LandscapeType]int
	LTCostFunctions map[nodetype.LandscapeType]func(node.Node, grid.AccessibilityGridStruct)
	CostFunctions   map[nodetype.GroundType]func(node.Node, grid.AccessibilityGridStruct)
}

//Name of the generator
func (rg RoadGenerator) Name() string {
	return "RoadGenerator"
}

//Create a new road generator with randomized conf
func Create() (rg RoadGenerator) {
	rg.RoadInfluenceCost = -3

	rg.DefaultDepthReach = 3
	rg.DefaultDepthCost = 2

	rg.CostDepth = make(map[nodetype.GroundType]int)
	rg.CostDepth[nodetype.NoGround] = 0
	rg.CostDepth[nodetype.Desert] = 3
	rg.CostDepth[nodetype.Plain] = 1
	rg.CostDepth[nodetype.Sea] = 999

	rg.LTCostDepth = make(map[nodetype.LandscapeType]int)
	rg.LTCostDepth[nodetype.NoLandscape] = 0
	rg.LTCostDepth[nodetype.Forest] = 3
	rg.LTCostDepth[nodetype.Mountain] = 3
	rg.LTCostDepth[nodetype.River] = 15

	rg.CostReach = make(map[nodetype.LandscapeType]int)
	rg.CostReach[nodetype.Forest] = 2
	rg.CostReach[nodetype.Mountain] = 3
	rg.CostReach[nodetype.River] = 0

	rg.Connection = make(map[int]bool)
	rg.Connection[north] = true
	rg.Connection[east] = true
	rg.Connection[south] = true
	rg.Connection[west] = true

	rg.CostFunctions = make(map[nodetype.GroundType]func(node.Node, grid.AccessibilityGridStruct))
	rg.CostFunctions[nodetype.Desert] = rg.computeDefaultCost
	rg.CostFunctions[nodetype.Plain] = rg.computeDefaultCost
	rg.CostFunctions[nodetype.Sea] = rg.refuse

	rg.LTCostFunctions = make(map[nodetype.LandscapeType]func(node.Node, grid.AccessibilityGridStruct))
	rg.LTCostFunctions[nodetype.Forest] = rg.computeDefaultReachCost
	rg.LTCostFunctions[nodetype.River] = rg.computeRiverCost
	rg.LTCostFunctions[nodetype.Mountain] = rg.computeDefaultReachCost

	return
}

func (rg RoadGenerator) fetchLTCost(nt nodetype.LandscapeType) int {
	c, has := rg.LTCostDepth[nt]
	if has {
		return c
	}
	return rg.DefaultDepthCost
}

func (rg RoadGenerator) fetchCost(nt nodetype.GroundType) int {
	c, has := rg.CostDepth[nt]
	if has {
		return c
	}
	return rg.DefaultDepthCost
}

func (rg RoadGenerator) fetchReach(nt nodetype.LandscapeType) int {
	c, has := rg.CostReach[nt]
	if has {
		return c
	}
	return rg.DefaultDepthReach
}

func (rg RoadGenerator) computeDefaultReachCost(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.Apply(n.Location, pattern.GenerateAdjascentPattern(rg.fetchReach(n.Landscape)), func(nn *node.Node, data int) (newData int) { return data + rg.fetchLTCost(n.Landscape) })
}

func (rg RoadGenerator) computeDefaultCost(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, acc.GetData(n.Location)+rg.fetchCost(n.Ground))
}

func (rg RoadGenerator) computeRiverCost(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, acc.GetData(n.Location)+rg.fetchLTCost(nodetype.River))
}

func (rg RoadGenerator) refuse(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, 999)
}

//Level of the sub generator see Generator Level
func (rg RoadGenerator) Level() map_level.GeneratorLevel {
	return map_level.Transportation
}

func (rg RoadGenerator) computeCost(nd node.Node, acc grid.AccessibilityGridStruct) {
	cost, has := rg.CostFunctions[nd.Ground]
	if nd.IsRoad {
		acc.SetData(nd.Location, tools.Min(acc.GetData(nd.Location)-2, 0))
		acc.Apply(nd.Location, pattern.Adjascent, func(nn *node.Node, data int) (newData int) {
			if nn.IsRoad {
				return data
			}
			return data + 1
		})
	}
	if acc.IsAccessible(nd.Location) {
		if has {
			cost(nd, acc)
			cost, has = rg.LTCostFunctions[nd.Landscape]
			if has {
				cost(nd, acc)
			}
		}
	} else {
		rg.refuse(nd, acc)
	}
}

func (rg RoadGenerator) astarGrid(gd *grid.CompoundedGrid, tempGrid *grid.AccessibilityGridStruct, origin, target node.Point) {
	var current = make(map[int]node.Point)
	current[target.ToInt(gd.Base.Size)] = target
	var next = make(map[int]node.Point)

	currentDist := 0

	used := make(map[int]bool)

	for true {

		for _, v := range current {
			if !used[v.ToInt(gd.Base.Size)] {
				tempGrid.SetData(v, currentDist+tempGrid.GetData(v))
				used[v.ToInt(gd.Base.Size)] = true
				for _, w := range tempGrid.SelectPattern(v, pattern.Adjascent) {
					if !used[w.ToInt(gd.Base.Size)] {
						if _, ok := next[w.ToInt(gd.Base.Size)]; !ok {
							next[w.ToInt(gd.Base.Size)] = w
						}
					}
				}
			}
		}
		current = next
		currentDist++
		next = make(map[int]node.Point)
		if len(current) == 0 {
			break
		}
	}
	log.Printf("RG: Acc: \n%s", tempGrid.String())

}

type generatedRoad struct {
	Road   node.Path
	Nodes  map[int]bool
	Cities map[int]bool
}

//Generate Will apply generator to provided grid
func (rg RoadGenerator) Generate(gd *grid.CompoundedGrid, dbh *db.Handler) error {
	// prepare a road heat map.
	// rivers are a solid +10 in difficulty
	// mountains and forest stacks +2 for each depth of each
	// desert stacks +3 for each depth
	// Preexisting roads count for a solid - 15 in difficulty
	// Area near a road is easier to access.

	return nil

	cities := gd.Base.LocationToCity
	clist := make([]int, 0, len(cities))
	for k := range cities {
		clist = append(clist, k)
	}

	rand.Shuffle(len(clist), func(i, j int) { clist[i], clist[j] = clist[j], clist[i] })

	visitedCities := make(map[int]bool)

	roads := make([]generatedRoad, 0, 0)

	for _, k := range clist {
		originCity := cities[k]
		acc := gd.AccessibilityGrid()

		// on cherche les villes :) (on les met dans un array ? :) )
		for x := 0; x < gd.Base.Size; x++ {
			for y := 0; y < gd.Base.Size; y++ {
				nd := gd.GetP(x, y)
				rg.computeCost(nd, acc)
			}
		}

		visitedCities[k] = true

		var targetCity *city.City
		targetCity = nil

		for kk, vv := range cities {
			if !visitedCities[kk] {
				targetCity = vv
				visitedCities[kk] = true
				break
			}
		}

		if targetCity == nil {
			for kk, vv := range cities {
				if kk != k {

					targetCity = vv
					break
				}
			}
		}

		// on selection 2 villes (originCity, targetCity)
		rg.astarGrid(gd, &acc, originCity.Location, targetCity.Location)

		log.Printf("Acc Map: \n%s", acc.String())
		var gr generatedRoad
		gr.Cities = make(map[int]bool)
		gr.Nodes = make(map[int]bool)
		gr.Cities[originCity.ID] = true
		gr.Cities[targetCity.ID] = true
		gr.Road = make([]node.Point, 0)

		// on cherche le chemin le plus court et le moins cher
		currentLocation := originCity.Location
		target := targetCity.Location

		for currentLocation != target {
			// seek lowest point from current location.
			currentLowest := 999
			nbRoadMet := 0
			var currentPoint node.Point
			for _, targetNode := range acc.SelectPattern(currentLocation, pattern.Adjascent) {
				val := acc.GetData(targetNode)
				if gd.Get(targetNode).IsRoad {
					nbRoadMet++
				}

				if gr.Nodes[targetNode.ToInt(gd.Base.Size)] {
					// cant go backward
					continue
				}
				if val < currentLowest {
					currentLowest = val
					currentPoint = targetNode
				}
			}

			if currentLowest >= 999 || nbRoadMet > 2 {
				// go backward and mark this one at 999
				// go backward if there are already 3 ajd road to this one.
				acc.SetData(currentLocation, 999)
				delete(gr.Nodes, currentLocation.ToInt(gd.Base.Size))
				currentLocation = gr.Road[len(gr.Road)-1]
				gr.Road = gr.Road[:len(gr.Road)-1]
				continue
			}
			gr.Road = append(gr.Road, currentPoint)
			gr.Nodes[currentPoint.ToInt(gd.Base.Size)] = true
			currentLocation = currentPoint
		}
		// on stock qu'on a bien relier les deux villes au reseau.

		roads = append(roads, gr)

		// on met le flag road sur toute les cases.

		for _, rloc := range gr.Road {
			gd.SetPRoad(rloc.X, rloc.Y, true)
		}
		log.Printf("Map: \n%s", gd.Delta.String())
	}

	return nil
}
