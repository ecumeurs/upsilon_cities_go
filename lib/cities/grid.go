package cities

import (
	"time"
)

type Grid struct {
	ID         int
	Nodes      [][]Node
	Name       string
	LastUpdate time.Time
	Cities     []City
}

// create a new random grid.
func New() *Grid {
	grid := new(Grid)
	grid.ID = 0
	grid.LastUpdate = time.Now()

	// generate map ... size

	for {
		grid.Generate(10, 5)
		// grid has been generated randomly ... now clear out unwanted cities (those not matching)

		// is it still valid ? if so break, otherwise try again ;)
		if grid.IsValid() {
			break
		}
	}

	return grid
}

func (grid *Grid) String() string {
	var res string
	for _, line := range grid.Nodes {
		var resline string
		for _, node := range line {
			resline += node.Short()
		}
		res += resline + "\n"
	}
	return res
}

//Generate generate a new grid
func (grid *Grid) Generate(maxSize int, scarcity int) {
	grid.Nodes = nil
	currentID := 0
	for i := 0; i < maxSize; i++ {
		var line []Node
		for j := 0; j < maxSize; j++ {
			var node Node
			node.ID = currentID
			currentID++
			node.Location.X = j
			node.Location.Y = i
			node.RandomCity(scarcity)
			line = append(line, node)
		}
		grid.Nodes = append(grid.Nodes, line)
	}
}

//IsValid check grid validity
func (grid *Grid) IsValid() bool {
	return true
}
