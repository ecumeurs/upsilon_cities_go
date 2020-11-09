package road_generator

import (
	"log"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
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

func (rg RoadGenerator) computeCost(node node.Node, acc grid.AccessibilityGridStruct) {
	cost, has := rg.CostFunctions[node.Ground]
	if has {
		cost(node, acc)
		cost, has = rg.LTCostFunctions[node.Landscape]
		if has {
			cost(node, acc)
		} else {
			rg.refuse(node, acc)
		}
	} else {
		rg.refuse(node, acc)
	}
}

//Generate Will apply generator to provided grid
func (rg RoadGenerator) Generate(gd *grid.CompoundedGrid, dbh *db.Handler) error {
	// prepare a road heat map.
	// rivers are a solid +10 in difficulty
	// mountains and forest stacks +2 for each depth of each
	// desert stacks +3 for each depth
	// Preexisting roads count for a solid - 15 in difficulty
	// Area near a road is easier to access.

	acc := gd.AccessibilityGrid()

	for true {
		// on cherche les villes :) (on les met dans un array ? :) )
		for x := 0; x < gd.Base.Size; x++ {
			for y := 0; y < gd.Base.Size; y++ {
				rg.computeCost(gd.GetP(x, y), acc)
			}
		}

		log.Printf("Cost Map: \n%s", acc.String())

		// on selection 2 villes

		// on cherche le chemin le plus court et le moins cher

		// on met le flag road sur toute les cases.

		// on stock qu'on a bien relier les deux villes au reseau.

		// on verifie qu'on est pas arrivé a la fin ( toutes les villes reliées au reseau, 1 seul reseau )
		return nil
	}

	return nil
}
