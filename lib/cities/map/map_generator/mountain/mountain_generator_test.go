package mountain_generator

import (
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/node"
)

func TestMountainGenerator(t *testing.T) {
	mg := Create()
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, node.Plain)
	gd.Delta = grid.Create(20, node.None)

	mg.Generate(gd)

	//t.Error(gd.Delta.String())
}
