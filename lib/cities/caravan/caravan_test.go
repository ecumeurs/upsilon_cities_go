package caravan

import (
	"log"
	"testing"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/generator"
)

func prepare() (*db.Handler, *grid.Grid) {
	dbh := db.NewTest()
	db.FlushDatabase(dbh)

	tools.InitCycle()
	// ensure that in memory storage is fine.
	city_manager.InitManager()
	grid_manager.InitManager()

	generator.CreateSampleFile()
	generator.Init()

	producer.CreateSampleFile()
	producer.Load()

	return dbh, grid.New(dbh)
}

func getNeighbours(grd *grid.Grid) (*city.City, *city.City) {
	for _, v := range grd.Cities {
		for _, w := range v.NeighboursID {
			if grd.Cities[w].CorporationID != v.CorporationID {
				return v, grd.Cities[w]
			}
		}
	}
	log.Fatalf("Failed to find a neighbourhood")
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
	crv.MapID = grd.ID
	crv.Exported.ItemType = []string{"Iron"}
	crv.Exported.Quality.Min = 5
	crv.Exported.Quality.Max = 50
	crv.Exported.Quantity.Min = 5
	crv.Exported.Quantity.Max = 50

	crv.Imported.ItemType = []string{"Wood"}
	crv.Imported.Quality.Min = 5
	crv.Imported.Quality.Max = 50
	crv.Imported.Quantity.Min = 5
	crv.Imported.Quantity.Max = 50

	if !crv.IsValid() {
		t.Errorf("Should have been valid")
		log.Printf("Caravan: %+v", crv)
		log.Printf("Cities: %+v %+v", lhs, rhs)
		return
	}

	crv.Insert(dbh)

	lhs.Reload(dbh)
	rhs.Reload(dbh)

	clhs.Reload(dbh)
	crhs.Reload(dbh)

	if len(lhs.CaravanID) != 1 {
		t.Errorf("lhs city should have at least a caravan")
		log.Printf("LHS city caravans %+v", lhs)
		return
	}
	if len(rhs.CaravanID) != 1 {
		t.Errorf("rhs city should have at least a caravan")
		return
	}
	if len(clhs.CaravanID) != 1 {
		t.Errorf("lhs corporation should have at least a caravan")
		return
	}

	if lhs.CaravanID[0] != crv.ID {
		t.Errorf("lhs city should have the right caravan, but doesnt")
		return
	}

	if rhs.CaravanID[0] != crv.ID {
		t.Errorf("rhs city should have the right caravan, but doesnt")
		return
	}

	if clhs.CaravanID[0] != crv.ID {
		t.Errorf("lhs corporation should have the right caravan, but doesnt")
		return
	}

	if crhs.CaravanID[0] != crv.ID {
		t.Errorf("rhs corporation should have the right caravan, but doesnt")
		return
	}

	if crv.Store.Capacity < tools.Max(crv.Exported.Quantity.Max, crv.Imported.Quantity.Max) {
		t.Errorf("Caravan storage should have been at least of max import/export quantity")
		log.Printf("Caravan store capacity: %d exported %d imported %d", crv.Store.Capacity, crv.Exported.Quantity.Max, crv.Imported.Quantity.Max)
		return
	}
}

type testContext struct {
	dbh  *db.Handler
	clhs *corporation.Corporation
	crhs *corporation.Corporation
	lhs  *city.City
	rhs  *city.City
	grd  *grid.Grid
	crv  *Caravan
}

func generateCaravan() testContext {
	var tst testContext

	tst.dbh, tst.grd = prepare()

	tst.lhs, tst.rhs = getNeighbours(tst.grd)
	tst.clhs, _ = getCorpo(tst.dbh, tst.lhs)
	tst.crhs, _ = getCorpo(tst.dbh, tst.rhs)

	tst.crv = New()
	tst.crv.CorpOriginID = tst.clhs.ID
	tst.crv.CorpTargetID = tst.crhs.ID
	tst.crv.CityOriginID = tst.lhs.ID
	tst.crv.CityTargetID = tst.rhs.ID
	tst.crv.MapID = tst.grd.ID
	tst.crv.Exported.ItemType = []string{"Iron"}
	tst.crv.Exported.Quality.Min = 5
	tst.crv.Exported.Quality.Max = 50
	tst.crv.Exported.Quantity.Min = 5
	tst.crv.Exported.Quantity.Max = 50

	tst.crv.Imported.ItemType = []string{"Wood"}
	tst.crv.Imported.Quality.Min = 5
	tst.crv.Imported.Quality.Max = 50
	tst.crv.Imported.Quantity.Min = 5
	tst.crv.Imported.Quantity.Max = 50
	err := tst.crv.Insert(tst.dbh)
	if err != nil {

		log.Printf("Caravan %+v", tst.crv)
		log.Fatalf("Err %s", err)
	}

	tst.lhs.Reload(tst.dbh)
	tst.rhs.Reload(tst.dbh)

	tst.clhs.Reload(tst.dbh)
	tst.crhs.Reload(tst.dbh)

	log.Printf("Caravan %+v", tst.crv)
	log.Printf("Cities %+v %+v", tst.lhs, tst.rhs)
	log.Printf("Corporation %+v %+v", tst.clhs, tst.crhs)

	return tst
}

func TestCaravanGetAccepted(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	err := tst.crv.Accept(tst.dbh, tst.crhs.ID)

	if err != nil {
		t.Errorf("should have been accepted.")
		return
	}

	if tst.crv.State != CRVWaitingOriginLoad {
		t.Errorf("state should have been waiting load")
		return
	}
}

func TestCaravanOnlyRHSCanAcceptCaravan(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	err := tst.crv.Accept(tst.dbh, tst.clhs.ID)

	if err == nil {
		t.Errorf("should not have been accepted . err %s", err)
		log.Printf("Caravan: %+v", tst.crv)
		log.Printf("Origin id: %d ; target id : %d", tst.clhs.ID, tst.crhs.ID)
		return
	}
}

func TestCaravanGetRefused(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	err := tst.crv.Refuse(tst.dbh, tst.crhs.ID)

	if err != nil {
		t.Errorf("should have been accepted.")
		log.Printf("Caravan: %+v", tst.crv)
		log.Printf("Origin id: %d ; target id : %d", tst.clhs.ID, tst.crhs.ID)

		return
	}
}

func TestCaravanOnlyRHSCanRefuseCaravan(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	err := tst.crv.Refuse(tst.dbh, tst.clhs.ID)

	if err == nil {
		t.Errorf("should not have been refused. err %s", err)
		log.Printf("Caravan: %+v", tst.crv)
		log.Printf("Origin id: %d ; target id : %d", tst.clhs.ID, tst.crhs.ID)

		return
	}
}
func TestCaravanCounterPropositionGetIssued(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	tst.crv.ExchangeRateLHS = 3
	err := tst.crv.Counter(tst.dbh, tst.crhs.ID)

	if err != nil {
		t.Errorf("should have been counter")
		return
	}

	if tst.crv.State != CRVCounterProposal {

		t.Errorf("should have been in counter proposal state")
		return
	}
}

func TestCaravanCounterPropositionGetAccepted(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()
	tst.crv.ExchangeRateLHS = 3
	err := tst.crv.Counter(tst.dbh, tst.crhs.ID)

	if err != nil {
		t.Errorf("should have been countered.")
		log.Printf("Caravan %+v", tst.crv)
		return
	}

	err = tst.crv.Accept(tst.dbh, tst.clhs.ID)

	if err != nil {
		t.Errorf("should have been accepted.")
		return
	}

	if tst.crv.State != CRVWaitingOriginLoad {
		t.Errorf("state should have been waiting load")
		return
	}
}

func TestCaravanCounterPropositionOnlyLHSCanAcceptCaravan(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	tst.crv.ExchangeRateLHS = 3
	err := tst.crv.Counter(tst.dbh, tst.crhs.ID)
	if err != nil {
		t.Errorf("should have been countered. %s", err)
		return
	}
	err = tst.crv.Accept(tst.dbh, tst.crhs.ID)

	if err == nil {
		t.Errorf("should not have been accepted. err %s", err)
		return
	}
}

func TestCaravanCounterPropositionGetRefused(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	tst.crv.ExchangeRateLHS = 3
	err := tst.crv.Counter(tst.dbh, tst.crhs.ID)
	err = tst.crv.Refuse(tst.dbh, tst.clhs.ID)

	if err != nil {
		t.Errorf("should have been accepted.")
		return
	}
}

func TestCaravanCounterPropositionOnlyRHSCanRefuseCaravan(t *testing.T) {
	tst := generateCaravan()
	defer tst.dbh.Close()

	tst.crv.ExchangeRateLHS = 3
	err := tst.crv.Counter(tst.dbh, tst.crhs.ID)
	if err != nil {
		t.Errorf("should have been countered. err %s", err)
		return
	}
	err = tst.crv.Refuse(tst.dbh, tst.crhs.ID)

	if err == nil {
		t.Errorf("should not have been refused. err %s", err)
		return
	}
}

func generateValidCaravan() testContext {
	tst := generateCaravan()
	err := tst.crv.Accept(tst.dbh, tst.crhs.ID)

	if err != nil {
		log.Fatalf("Caravan: Should have been accepted %s %v+", err, tst.crv)
	}

	return tst
}

func TestCaravanGetAborted(t *testing.T) {
	tst := generateValidCaravan()
	defer tst.dbh.Close()
	tst.crv.Abort(tst.dbh)

	if tst.crv.IsAborted() {
		t.Errorf("caravan state should have been aborted.")
		log.Printf("Caravan: %+v", tst.crv)
		log.Printf("Origin id: %d ; target id : %d", tst.clhs.ID, tst.crhs.ID)
	}
}

func TestCaravanGetLoaded(t *testing.T) {
	tst := generateValidCaravan()
	defer tst.dbh.Close()

	// origin is expected to export iron.
	//
	//crv.Exported.BasePrice = 10
	//crv.Exported.ItemType = []string{"Iron"}
	//crv.Exported.Quality.Min = 5
	//crv.Exported.Quality.Max = 50
	//crv.Exported.Quantity.Min = 5
	//crv.Exported.Quantity.Max = 50
	//crv.Exported.BasePrice = 10
	//crv.Exported.BasePrice = 10

	// forcefully add iron to the city store...
	// caravan should see it's storage updated instead.
	var it item.Item
	it.BasePrice = 15 // has be renegotiated that's all
	it.Type = []string{"Iron"}
	it.Name = "Iron Ingot"
	it.Quality = 13
	it.Quantity = 65 // more than expected, city storage should have only 15 left after this operation.

	log.Printf("Adding iron to city's store")
	tst.lhs.Storage.Add(it)

	log.Printf("Filling caravan form city's store")
	err := tst.crv.Fill(tst.dbh, tst.lhs)
	if err != nil {
		t.Errorf("caravan should have been filled.")
		return
	}

	log.Printf("Checking city store")
	items := tst.lhs.Storage.All(storage.ByType("Iron"))
	if len(items) != 1 {
		t.Errorf("city storage doesn't have iron")
		return
	}
	if items[0].Quantity != 15 {
		t.Errorf("city storage should still have 15 iron")
		return
	}

	// might not have same pointer in city's storage ... so reload ;)
	tst.crv.Reload(tst.dbh)
	log.Printf("Checking caravan store after reload")

	items = tst.crv.Store.All(storage.ByType("Iron"))
	if len(items) != 1 {
		t.Errorf("caravan storage doesn't have iron")
		return
	}

	if items[0].Quantity != 50 {
		t.Errorf("caravan storage should  have 50 iron")
		return
	}

	if !tst.crv.IsFilled() {
		t.Errorf("caravan should be filled and ready to go.")
		log.Printf("Caravan: %+v", tst.crv)
		return
	}

}

func TestCaravanGetLoadedMultipleTimes(t *testing.T) {
	tst := generateValidCaravan()
	defer tst.dbh.Close()

	// origin is expected to export iron.
	//
	//crv.Exported.BasePrice = 10
	//crv.Exported.ItemType = []string{"Iron"}
	//crv.Exported.Quality.Min = 5
	//crv.Exported.Quality.Max = 50
	//crv.Exported.Quantity.Min = 5
	//crv.Exported.Quantity.Max = 50
	//crv.Exported.BasePrice = 10
	//crv.Exported.BasePrice = 10

	// forcefully add iron to the city store...
	// caravan should see it's storage updated instead.
	var it item.Item
	it.BasePrice = 15 // has be renegotiated that's all
	it.Type = []string{"Iron"}
	it.Name = "Iron Ingot"
	it.Quality = 13
	it.Quantity = 5

	tst.lhs.Storage.Add(it)

	err := tst.crv.Fill(tst.dbh, tst.lhs)
	if err != nil {
		t.Errorf("caravan should have been filled.")
		return
	}

	if tst.crv.IsFilled() {
		t.Errorf("caravan shouldn't be filled")
		log.Printf("Caravan: %+v", tst.crv)
		return
	}

	it.Quantity = 30

	tst.lhs.Storage.Add(it)

	err = tst.crv.Fill(tst.dbh, tst.lhs)
	if err != nil {
		t.Errorf("caravan should have been filled.")
		return
	}

	if tst.crv.IsFilled() {
		t.Errorf("caravan shouldn't be filled")
		log.Printf("Caravan: %+v", tst.crv)
		return
	}

	it.Quantity = 30

	tst.lhs.Storage.Add(it)

	err = tst.crv.Fill(tst.dbh, tst.lhs)
	if err != nil {
		t.Errorf("caravan should have been filled.")
		return
	}

	if !tst.crv.IsFilled() {
		t.Errorf("caravan should be filled")
		log.Printf("Caravan: %+v", tst.crv)
		return
	}

	log.Printf("Checking city store")
	items := tst.lhs.Storage.All(storage.ByType("Iron"))
	if len(items) != 1 {
		t.Errorf("city storage doesn't have iron")
		return
	}
	if items[0].Quantity != 15 {
		t.Errorf("city storage should still have 15 iron")
		return
	}

	it.Quantity = 5

	tst.lhs.Storage.Add(it)

	err = tst.crv.Fill(tst.dbh, tst.lhs)
	if err != nil {
		t.Errorf("caravan should have been filled.")
		return
	}

	if !tst.crv.IsFilled() {
		t.Errorf("caravan should be filled")
		log.Printf("Caravan: %+v", tst.crv)
		return
	}

	log.Printf("Checking city store")
	items = tst.lhs.Storage.All(storage.ByType("Iron"))
	if len(items) != 1 {
		t.Errorf("city storage doesn't have iron")
		return
	}
	if items[0].Quantity != 20 {
		t.Errorf("city storage should still have 15 iron")
		return
	}

	items = tst.crv.Store.All(storage.ByType("Iron"))
	if len(items) != 1 {
		t.Errorf("caravan storage doesn't have iron")
		return
	}

	if items[0].Quantity != 50 {
		t.Errorf("caravan storage should not have more than 50 iron")
		return
	}

}

func TestCaravanGetPartiallyLoaded(t *testing.T) {

	tst := generateValidCaravan()
	defer tst.dbh.Close()

	// caravan should see it's storage updated instead.
	var it item.Item
	it.BasePrice = 15 // has be renegotiated that's all
	it.Type = []string{"Iron"}
	it.Name = "Iron Ingot"
	it.Quality = 13
	it.Quantity = 35 // less than expected...

	tst.lhs.Storage.Add(it)

	err := tst.crv.Fill(tst.dbh, tst.lhs)
	if err != nil {
		t.Errorf("caravan should have been filled.")
		return
	}

	items := tst.lhs.Storage.All(storage.ByType("Iron"))
	if len(items) != 0 {
		t.Errorf("city storage shouldn't have iron")
		return
	}

	// might not have same pointer in city's storage ... so reload ;)
	tst.crv.Reload(tst.dbh)

	items = tst.crv.Store.All(storage.ByType("Iron"))
	if len(items) != 1 {
		t.Errorf("caravan storage doesn't have iron")
		return
	}
	if items[0].Quantity != 35 {
		t.Errorf("caravan storage should  have 35 iron")
		return
	}
	if tst.crv.IsFilled() {
		t.Errorf("caravan shouldn't be filled and ready to go")
		return
	}
	if !tst.crv.IsFilledAtAcceptableLevel() {
		t.Errorf("caravan should be acceptably filled and ready to go")
		return
	}
}

func generateFilledCaravan() testContext {

	tst := generateValidCaravan()

	// origin is expected to export iron.
	//
	//crv.Exported.BasePrice = 10
	//crv.Exported.ItemType = []string{"Iron"}
	//crv.Exported.Quality.Min = 5
	//crv.Exported.Quality.Max = 50
	//crv.Exported.Quantity.Min = 5
	//crv.Exported.Quantity.Max = 50
	//crv.Exported.BasePrice = 10
	//crv.Exported.BasePrice = 10

	// forcefully add iron to the city store...
	// caravan should see it's storage updated instead.
	var it item.Item
	it.BasePrice = 15 // has be renegotiated that's all
	it.Type = []string{"Iron"}
	it.Name = "Iron Ingot"
	it.Quality = 13
	it.Quantity = 65 // more than expected, city storage should have only 15 left after this operation.

	tst.lhs.Storage.Add(it)

	tst.crv.Fill(tst.dbh, tst.lhs)

	// might not have same pointer in city's storage ... so reload ;)
	tst.crv.Reload(tst.dbh)

	return tst
}

func TestCaravanGetTravelsToTarget(t *testing.T) {
	tst := generateFilledCaravan()
	defer tst.dbh.Close()

	rnd := tools.RoundTime(time.Now().UTC())
	done, err := tst.crv.TimeToMove(tst.dbh, tst.lhs, rnd)

	if err != nil {
		t.Errorf("should have successfully be set to move %s", err)
		return
	}

	if !done {
		t.Errorf("now should have been good time to move ... but it isn't %+v", tst.crv)
		return
	}

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

func TestCaravanGetTerminatedByEndOfTerm(t *testing.T) {
	t.Errorf("Not Implemented")

}
