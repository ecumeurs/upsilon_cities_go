package region

import (
	"testing"
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

	_, err = reg.Generate(dbh)
	if err != nil {
		t.Errorf("Failed to generate a grid based on Elvenwood region: %s", err)
		return
	}
}
