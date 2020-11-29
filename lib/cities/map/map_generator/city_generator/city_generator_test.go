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
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/system"
	"upsilon_cities_go/lib/misc/generator"
)

func TestCityGenerator(t *testing.T) {
	system.LoadConf()
	generator.Load()
	dbh := db.NewTest()
	db.FlushDatabase(dbh)

	dg := Create()
	dg.Density.Min = 3
	dg.Density.Max = 3
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Delta = grid.Create(20, nodetype.NoGround)
	gd.Base.Insert(dbh)

	resource_generator.Load()
	for idx := range gd.Base.Nodes {
		gd.Base.Nodes[idx].Activated = []resource.Resource{resource_generator.MustOne("Fer")}
	}

	dg.Generate(gd, dbh)

	if len(gd.Delta.Cities) == 0 {
		t.Error("Failed ! Expected cities to have been generated.")
	}

	if len(gd.Delta.LocationToCity) == 0 {
		t.Error("Failed ! Expected cities to have been generated and quick reference map as well.")
	}

	if len(gd.Delta.Cities) != 6 {
		t.Errorf("Failed ! Expected 6 cities to have been generated, got %d.", len(gd.Delta.Cities))
	}

	if len(gd.Delta.LocationToCity) != len(gd.Delta.Cities) {
		t.Error("Failed ! Expected location to cities to be filled with all cities.")
	}

	for _, v := range gd.Delta.Cities {

		if v.ID == 0 {
			t.Error("Expected city to have an id.")
			return
		}

		if v.MapID == 0 {
			t.Error("Expected city to have a map id")
			return
		}

		if v.Location.IsEq(node.NP(0, 0)) {
			t.Error("Expected city to have a location")
		}

		ncty, hasValue := gd.Delta.LocationToCity[v.Location.ToInt(gd.Base.Size)]

		if hasValue == false {
			t.Error("Expected city to have been registered in location map.")
		}

		if v.ID != ncty.ID {
			t.Error("Expected city at location have the same id")
		}
	}

}

func TestGenerateCityCreation(t *testing.T) {
	system.LoadConf()
	generator.Load()
	dbh := db.NewTest()
	db.FlushDatabase(dbh)
	dg := Create()
	dg.Density.Min = 3
	dg.Density.Max = 3
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Base.Insert(dbh)

	resource_generator.Load()

	for idx := range gd.Base.Nodes {
		gd.Base.Nodes[idx].Activated = []resource.Resource{resource_generator.MustOne("Fer")}
	}
	gd.Delta = grid.Create(20, nodetype.NoGround)

	dg.generateCityPrepare(gd, dbh, node.NP(10, 10))

	// expect a city to have been added to the stack !

	if len(gd.Delta.Cities) == 0 {
		t.Error("Expected a city to have been generated")
		return // can't continue tests ... that one was mandatory ;)
	}
}

func TestGenerateCityFilling(t *testing.T) {

	system.LoadConf()
	dbh := db.NewTest()
	db.FlushDatabase(dbh)

	dg := Create()
	dg.Density.Min = 3
	dg.Density.Max = 3
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Base.Insert(dbh)

	resource_generator.Load()
	producer_generator.Load()
	gd.Delta = grid.Create(20, nodetype.NoGround)

	for idx := range gd.Base.Nodes {
		gd.Base.Nodes[idx].Activated = []resource.Resource{resource_generator.MustOne("Fer")}
	}

	dg.generateCity(gd, dbh, node.NP(10, 10))

	// expect a city to have been added to the stack !

	if len(gd.Delta.Cities) == 0 {
		t.Error("Expected a city to have been generated")
		return // can't continue tests ... that one was mandatory ;)
	}

	n := gd.Delta.Get(node.NP(10, 10))
	if !n.IsStructure {
		t.Errorf("Expected Delta node to have been marked as Structure: %v", n)
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

	gd.Compact()

	n = gd.Base.Get(node.NP(10, 10))

	if !n.IsStructure {
		t.Error("Expected city to be available in compacted grid.")
		t.Errorf("Node %v", n)
	}
}
