package resource_generator

import (
	"testing"
	rg "upsilon_cities_go/lib/cities/city/resource_generator"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/forest_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/mountain_generator"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/system"
)

func TestResourceGenerator(t *testing.T) {

	system.LoadConf()
	dbh := db.NewTest()
	db.FlushDatabase(dbh)
	rg.Load()

	mg := mountain_generator.Create()
	fg := forest_generator.Create()

	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)
	gd.Delta = grid.Create(20, nodetype.NoGround)

	mg.Generate(gd, dbh)
	gd.Base = gd.Compact()
	gd.Delta = grid.Create(20, nodetype.NoGround)
	fg.Generate(gd, dbh)
	gd.Base = gd.Compact()
	gd.Delta = grid.Create(20, nodetype.NoGround)

	rcg := Create()
	rcg.Generate(gd, dbh)

	for _, nd := range gd.Delta.Nodes {
		a := len(nd.Activated)
		p := len(nd.Potential)
		if a+p == 0 {
			t.Errorf("Should at least have something available as resources on each node. %v", nd.Location.String())
			return
		}
	}

	gd.Base = gd.Compact()

	for _, nd := range gd.Base.Nodes {

		a := len(nd.Activated)
		p := len(nd.Potential)
		if a+p == 0 {
			t.Errorf("Compact should keep activated resources and potential as well ...")
			return
		}
	}

}
