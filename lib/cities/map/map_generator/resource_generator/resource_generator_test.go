package resource_generator

import (
	"testing"
	rg "upsilon_cities_go/lib/cities/city/resource_generator"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/forest_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/mountain_generator"
	"upsilon_cities_go/lib/cities/nodetype"
)

func TestResourceGenerator(t *testing.T) {

	rg.Load()

	mg := mountain_generator.Create()
	fg := forest_generator.Create()

	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Delta = grid.Create(20, nodetype.None)

	mg.Generate(gd)
	gd.Base = gd.Compact()
	fg.Generate(gd)
	gd.Base = gd.Compact()

	rcg := Create()
	rcg.Generate(gd)

	t.Errorf(gd.Delta.String())
}
