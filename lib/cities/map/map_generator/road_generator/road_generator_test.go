package road_generator

import (
	"log"
	_ "net/http/pprof"
	"testing"
	"upsilon_cities_go/lib/cities/city/resource"
	"upsilon_cities_go/lib/cities/city/resource_generator"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/city_generator"
	"upsilon_cities_go/lib/cities/nodetype"
)

func TestRoadGenerator(t *testing.T) {

	dg := city_generator.Create()
	dg.Density.Min = 3
	dg.Density.Max = 3
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Delta = grid.Create(20, nodetype.None)

	for idx := range gd.Base.Nodes {
		gd.Base.Nodes[idx].Activated = []resource.Resource{resource_generator.MustOne("Fer")}
	}

	dg.Generate(gd)
	gd.Base = gd.Compact()
	gd.Delta = grid.Create(20, nodetype.None)

	rg := Create()
	rg.Generate(gd)
	gd.Base = gd.Compact()

	log.Printf("Delta: \n%s", gd.Delta.String())
	log.Printf("Result: \n%s", gd.Base.String())

	t.Error("Not implemented")
}
