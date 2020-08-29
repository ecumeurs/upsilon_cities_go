package city_generator

import (
	"testing"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city/producer_generator"
	"upsilon_cities_go/lib/cities/city/resource"
	"upsilon_cities_go/lib/cities/city/resource_generator"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
)

func TestCityGenerator(t *testing.T) {
	dg := Create()
	dg.Density.Min = 3
	dg.Density.Max = 3
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Delta = grid.Create(20, nodetype.None)

	for idx := range gd.Base.Nodes {
		gd.Base.Nodes[idx].Activated = []resource.Resource{resource_generator.MustOne("Fer")}
	}

	dg.Generate(gd)

	if len(gd.Delta.Cities) == 0 {
		t.Error("Failed ! Expected cities to have been generated.")
	}

	if len(gd.Delta.LocationToCity) == 0 {
		t.Error("Failed ! Expected cities to have been generated and quick reference map as well.")
	}

	if len(gd.Delta.Cities) != 6 {
		t.Error("Failed ! Expected 6 cities to have been generated.")
	}
}

func TestGenerateCityCreation(t *testing.T) {
	dg := Create()
	dg.Density.Min = 3
	dg.Density.Max = 3
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)

	resource_generator.Load()

	for idx := range gd.Base.Nodes {
		gd.Base.Nodes[idx].Activated = []resource.Resource{resource_generator.MustOne("Fer")}
	}
	gd.Delta = grid.Create(20, nodetype.None)

	dg.generateCityPrepare(gd, node.NP(10, 10))

	// expect a city to have been added to the stack !

	if len(gd.Delta.Cities) == 0 {
		t.Error("Expected a city to have been generated")
		return // can't continue tests ... that one was mandatory ;)
	}

	var cty *city.City

	for _, v := range gd.Delta.Cities {
		cty = v
		break
	}

	if cty == nil {
		t.Error("Expected to have a city: very weird")
		return
	}

	cty, hasValue := gd.Delta.LocationToCity[cty.Location.ToInt(gd.Base.Size)]

	if hasValue == false {
		t.Error("Expected city to have been registered in location map.")
	}
}

func TestGenerateCityFilling(t *testing.T) {
	dg := Create()
	dg.Density.Min = 3
	dg.Density.Max = 3
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)

	resource_generator.Load()
	producer_generator.Load()
	gd.Delta = grid.Create(20, nodetype.None)

	for idx := range gd.Base.Nodes {
		gd.Base.Nodes[idx].Activated = []resource.Resource{resource_generator.MustOne("Fer")}
	}

	dg.generateCity(gd, node.NP(10, 10))

	// expect a city to have been added to the stack !

	if len(gd.Delta.Cities) == 0 {
		t.Error("Expected a city to have been generated")
		return // can't continue tests ... that one was mandatory ;)
	}
	var cty *city.City

	for _, v := range gd.Delta.Cities {
		cty = v
		break
	}

	if cty == nil {
		t.Error("Expected to have a city: very weird")
		return
	}

	if len(cty.RessourceProducers) == 0 {
		t.Error("Expected to have at least one resource producer")

	}

	if len(cty.ProductFactories) == 0 {
		t.Error("Expected to have at least one factory")
	}
}
