package sea_generator

import (
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/node"
)

func TestMountainGenerator(t *testing.T) {
	sg := Create()
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, node.Plain)
	gd.Delta = grid.Create(20, node.None)

	sg.Generate(gd)

}
