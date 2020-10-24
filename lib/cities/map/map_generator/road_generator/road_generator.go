package road_generator

import (
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
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

	Connection    map[int]bool
	CostDepth     map[nodetype.NodeType]int
	CostReach     map[nodetype.NodeType]int
	CostFunctions map[nodetype.NodeType]func(*node.Node, grid.AccessibilityGridStruct)
}

//Create a new road generator with randomized conf
func Create() (mg RoadGenerator) {
	mg.RoadInfluenceCost = -3

	mg.DefaultDepthReach = 3
	mg.DefaultDepthCost = 2

	mg.CostDepth = make(map[nodetype.NodeType]int)
	mg.CostDepth[nodetype.Desert] = 3
	mg.CostDepth[nodetype.Plain] = 1
	mg.CostDepth[nodetype.River] = 15
	mg.CostDepth[nodetype.River] = 999
	mg.CostDepth[nodetype.Road] = -15

	mg.CostReach = make(map[nodetype.NodeType]int)
	mg.CostReach[nodetype.Plain] = 0
	mg.CostDepth[nodetype.River] = 0

	mg.Connection = make(map[int]bool)
	mg.Connection[north] = true
	mg.Connection[east] = true
	mg.Connection[south] = true
	mg.Connection[west] = true

	mg.CostFunctions = make(map[nodetype.NodeType]func(*node.Node, grid.AccessibilityGridStruct))
	mg.CostFunctions[nodetype.Desert] = mg.computeDefaultCost
	mg.CostFunctions[nodetype.Forest] = mg.computeDefaultCost
	mg.CostFunctions[nodetype.River] = mg.computeRiverCost
	mg.CostFunctions[nodetype.Plain] = mg.computeDefaultCost
	mg.CostFunctions[nodetype.Mountain] = mg.computeDefaultCost
	mg.CostFunctions[nodetype.Sea] = mg.refuse
	mg.CostFunctions[nodetype.CityNode] = mg.refuse
	mg.CostFunctions[nodetype.Road] = mg.refuse
	return
}

func (mg RoadGenerator) fetchCost(nt nodetype.NodeType) int {
	c, has := mg.CostDepth[nt]
	if has {
		return c
	}
	return mg.DefaultDepthCost
}

func (mg RoadGenerator) fetchReach(nt nodetype.NodeType) int {
	c, has := mg.CostReach[nt]
	if has {
		return c
	}
	return mg.DefaultDepthReach
}

func (mg RoadGenerator) computeDefaultCost(n *node.Node, acc grid.AccessibilityGridStruct) {
	acc.Apply(n.Location, pattern.GenerateAdjascentPattern(mg.fetchReach(n.Type)), func(nn *node.Node, data int) (newData int) { return data + mg.fetchCost(n.Type) })
}

func (mg RoadGenerator) computeRiverCost(n *node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, acc.GetData(n.Location)+mg.fetchCost(n.Type))
}

func (mg RoadGenerator) refuse(n *node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, 999)
}

//Level of the sub generator see Generator Level
func (mg RoadGenerator) Level() map_level.GeneratorLevel {
	return map_level.Transportation
}

func (mg RoadGenerator) computeCost(node *node.Node, acc grid.AccessibilityGridStruct) {
	cost, has := mg.CostFunctions[node.Type]
	if has {
		cost(node, acc)
	} else {
		mg.refuse(node, acc)
	}
}

//Generate Will apply generator to provided grid
func (mg RoadGenerator) Generate(gd *grid.CompoundedGrid) error {
	// prepare a road heat map.
	// rivers are a solid +10 in difficulty
	// mountains and forest stacks +2 for each depth of each
	// desert stacks +3 for each depth
	// Preexisting roads count for a solid - 15 in difficulty
	// Area near a road is easier to access.

	acc := gd.AccessibilityGrid()

	for x := 0; x < gd.Base.Size; x++ {
		for y := 0; y < gd.Base.Size; y++ {
			mg.computeCost(acc.GetP(x, y), acc)
		}
	}

	return nil
}
