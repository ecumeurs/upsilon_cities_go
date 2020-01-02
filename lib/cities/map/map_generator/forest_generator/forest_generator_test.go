package forest_generator

import (
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/node"
)

func TestForestGenerator(t *testing.T) {
	fg := Create()
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, node.Plain)
	gd.Delta = grid.Create(20, node.None)

	fg.Generate(gd)

}
