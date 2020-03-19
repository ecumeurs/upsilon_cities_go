package resource_generator

import (
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/forest_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/mountain_generator"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
)

func TestGatherResources(t *testing.T) {

	Load()

	mg := mountain_generator.Create()
	fg := forest_generator.Create()

	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Delta = grid.Create(20, nodetype.None)

	mg.Generate(gd)
	gd.Base = gd.Compact()
	fg.Generate(gd)
	gd.Base = gd.Compact()

	gd.Base.Get(node.NP(15, 4)).Type = nodetype.Plain

	depths := make(map[int]int)
	found := GatherResourcesAvailable(node.NP(15, 4), gd, &depths)
	if len(found) == 0 {
		t.Errorf("Failed to find resources for point 15,4")

	}

}
