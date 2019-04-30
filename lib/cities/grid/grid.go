package grid

import (
	"math/rand"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

type Grid struct {
	ID         int
	Nodes      []node.Node
	Name       string
	LastUpdate time.Time
	Cities     []city.City
	Size       int
}

// create a new random grid.
func New() *Grid {
	grid := new(Grid)
	grid.ID = 0
	grid.LastUpdate = time.Now()

	// generate map ... size

	grid.Generate(20, 5)
	// grid has been generated randomly ... now clear out unwanted cities (those not matching)
	grid.BuildRoad()

	return grid
}

func (grid *Grid) String() string {
	var res string
	i := 0
	for _, node := range grid.Nodes {
		res += node.Short()
		i++
		if i == grid.Size {
			res += "\n"
			i = 0
		}
	}
	return res
}

//Adapt will attempt to alter the grid so that it matches some rules
func (grid *Grid) BuildRoad() {

}

//Generate generate a new grid
func (grid *Grid) Generate(maxSize int, scarcity int) {
	grid.Nodes = nil
	grid.Size = maxSize
	currentID := 0
	for i := 0; i < maxSize; i++ {

		for j := 0; j < maxSize; j++ {
			var nde node.Node
			nde.ID = currentID
			currentID++
			nde.Location.X = j
			nde.Location.Y = i
			nde.Type = grid.RandomCity(nde.Location, scarcity)
			grid.Nodes = append(grid.Nodes, nde)
		}
	}
}

//Get will seek out a node.
func (grid *Grid) Get(location node.Point) *node.Node {
	if location.X > grid.Size-1 {
		return nil
	}
	if location.Y > grid.Size-1 {
		return nil
	}
	if grid.Size*location.Y+location.X >= len(grid.Nodes) {
		return nil
	}
	return &grid.Nodes[grid.Size*location.Y+location.X]
}

//GetP will seek out a node.
func (grid *Grid) GetP(x int, y int) *node.Node {
	if !tools.InEq(x, 0, grid.Size-1) {
		return nil
	}
	if !tools.InEq(y, 0, grid.Size-1) {
		return nil
	}
	if grid.Size*y+x >= len(grid.Nodes) {
		return nil
	}
	return &grid.Nodes[grid.Size*y+x]
}

//GetRange fetch nodes in range.
func (grid *Grid) GetRange(location node.Point, reach int) []*node.Node {
	location.X = tools.Max(0, location.X-reach/2)
	location.Y = tools.Max(0, location.Y-reach/2)

	var res []*node.Node

	for i := 0; i < reach; i++ {
		for j := 0; j < reach; j++ {
			pt := grid.GetP(location.X+j, location.Y+i)
			if pt != nil {
				res = append(res, pt)
			}
		}
	}
	return res
}

//RandomCity assign a random city; the higher scarcity the lower the chance to have a city ;)
func (grid *Grid) RandomCity(location node.Point, scarcity int) node.NodeType {
	roll := rand.Intn(scarcity + 1)
	if roll < scarcity {
		return node.None
	} else {
		// seek target location and a nice square of 3
		// if no cities are present in there then try it

		interloppers := grid.GetRange(location, 6)
		for _, nd := range interloppers {
			if nd.Type == node.CityNode {
				return node.None
			}
		}

		return node.CityNode

	}
}

//IsValid check grid validity
func (grid *Grid) IsValid() bool {
	return true
}
