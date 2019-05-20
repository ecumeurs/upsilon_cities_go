package caravan

import (
	"testing"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/db"
)

func prepare() (*db.Handler, *grid.Grid) {
	dbh := db.NewTest()
	db.FlushDatabase(dbh)

	return dbh, grid.New(dbh)
}

func getNeighbours(grd *grid.Grid) (*city.City, *city.City) {
	for _, v := range grd.Cities {
		return v, grd.Cities[v.NeighboursID[0]]
	}
	return nil, nil
}

func getCorpo(dbh *db.Handler, cty *city.City) (*corporation.Corporation, error) {
	return corporation.ByID(dbh, cty.CorporationID)
}

func TestCaravanGetProposed(t *testing.T) {
	dbh, grd := prepare()
	defer dbh.Close()
	lhs, rhs := getNeighbours(grd)
	clhs, _ := getCorpo(dbh, lhs)
	crhs, _ := getCorpo(dbh, rhs)

	crv := New()
	crv.CorpOriginID = clhs.ID
	crv.CorpTargetID = crhs.ID
	crv.CityOriginID = lhs.ID
	crv.CityTargetID = rhs.ID
	crv.Exported.BasePrice = 10
	crv.Exported.ItemType = []string{"Iron"}
	crv.Exported.Quality.Min = 5
	crv.Exported.Quality.Max = 50
	crv.Exported.Quantity.Min = 5
	crv.Exported.Quantity.Max = 50
	crv.Exported.BasePrice = 10
	crv.Exported.BasePrice = 10

	crv.Imported.ItemType = []string{"Wood"}
	crv.Imported.Quality.Min = 5
	crv.Imported.Quality.Max = 50
	crv.Imported.Quantity.Min = 5
	crv.Imported.Quantity.Max = 50
	crv.Imported.BasePrice = 10

	if !crv.IsValid() {
		t.Errorf("Should have been valid")
		return
	}

	crv.Insert(dbh)

	// caravans := crhs.FetchCaravans(dbh)

	// if len(caravans) == 0 {
	// 	t.Errorf("Should have found a caravan proposal on rhs")
	// 	return
	// }

	// if caravans[0].State != CRVProposal {
	// 	t.Errorf("Caravan should have been in proposal state.")
	// 	return
	// }

	// caravans := clhs.FetchCaravans(dbh)

	// if len(caravans) == 0 {
	// 	t.Errorf("Should have found a caravan proposal on lhs")
	// 	return
	// }

	// if caravans[0].State != CRVProposal {
	// 	t.Errorf("Caravan should have been in proposal state.")
	// 	return
	// }

}

func TestCaravanGetAccepted(t *testing.T) {

	t.Errorf("Not Implemented")
}

func TestCaravanGetRefused(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetAborted(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetTerminated(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetLoaded(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetPartiallyLoaded(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetTravelsToTarget(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetUnloadsTarget(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetLoadsForOrigin(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetPartiallyLoadsForOrigin(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetTravelsToOrigin(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetUnloadsOrigin(t *testing.T) {
	t.Errorf("Not Implemented")

}
func TestCaravanGetTravelsRestarts(t *testing.T) {
	t.Errorf("Not Implemented")

}

func TestCaravanGetAbortedByBrokenContract(t *testing.T) {
	t.Errorf("Not Implemented")

}

func TestCaravanGetAbortedByWillOfContractor(t *testing.T) {
	t.Errorf("Not Implemented")

}

func TestCaravanGetTerminatedByEndOfTerm(t *testing.T) {
	t.Errorf("Not Implemented")

}
