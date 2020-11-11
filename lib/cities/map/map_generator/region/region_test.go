package region

import (
	"testing"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/system"
)

func TestGenerateRegions(t *testing.T) {
	Load()
}

func TestGenerateOneRegion(t *testing.T) {
	Load()

	system.LoadConf()
	dbh := db.NewTest()
	db.FlushDatabase(dbh)
	db.MarkSessionAsTest() // forcefully replace all db.New by db.NewTest

	defer dbh.Close()

	reg, err := Generate("Elvenwood")
	if err != nil {
		t.Errorf("Failed to generate Elvenwood: %s", err)
		return
	}

	gd, err := reg.Generate(dbh)
	if err != nil {
		t.Errorf("Failed to generate a grid based on Elvenwood region: %s", err)
		return
	}

	if len(gd.Cities) == 0 {
		t.Error("Expected grid to have cities")
		return
	}

	var cty *city.City
	for _, c := range gd.Cities {
		cty = c
		break
	}

	if cty.ID == 0 {
		t.Error("Expected city to have an id.")
		return
	}

	if cty.MapID == 0 {
		t.Error("Expected city to have a map id.")
		return
	}

	grd, err := grid.ByID(dbh, gd.ID)

	if err != nil {
		t.Errorf("Expected to have loaded grid appropriately... but got error: %s", err)
		return
	}

	if len(grd.Cities) == 0 {
		t.Error("Expected loaded grid to have cities")
		return
	}

	for _, c := range grd.Cities {
		cty = c
		break
	}

	if cty.ID == 0 {
		t.Error("Expected city to have an id.")
		return
	}

	if cty.MapID == 0 {
		t.Error("Expected city to have a map id.")
		return
	}
}
