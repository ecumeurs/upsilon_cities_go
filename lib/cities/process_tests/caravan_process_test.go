package process_tests

import (
	"log"
	"testing"
	"time"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/generator"
)

func TestFullFlowCaravan(t *testing.T) {
	dbh := db.NewTest()
	db.FlushDatabase(dbh)
	db.MarkSessionAsTest() // forcefully replace all db.New by db.NewTest
	defer dbh.Close()

	caravan.Init()

	tools.InitCycle()
	// ensure that in memory storage is fine.
	city_manager.InitManager()
	grid_manager.InitManager()

	caravan_manager.InitManager()
	corporation_manager.InitManager()

	generator.CreateSampleFile()
	generator.Init()

	producer.CreateSampleFile()
	producer.Load()

	tgrid := grid.New(dbh)
	grid_manager.GenerateGridHandler(tgrid)

	grd, _ := grid_manager.GetGridHandler(tgrid.ID)

	var lhs, rhs *city_manager.Handler

	for _, v := range grd.Get().Cities {
		found := false
		if v.CorporationID == 0 {
			log.Printf("Grid shouldn't provide cities without corporation ...")
			continue
		}
		for _, w := range v.NeighboursID {
			if grd.Get().Cities[w].CorporationID != v.CorporationID && grd.Get().Cities[w].CorporationID != 0 {
				lhs, _ = city_manager.GetCityHandler(v.ID)
				rhs, _ = city_manager.GetCityHandler(w)
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	clhs, _ := corporation_manager.GetCorporationHandler(lhs.Get().CorporationID)
	crhs, _ := corporation_manager.GetCorporationHandler(rhs.Get().CorporationID)

	tcrv := caravan.New()
	tcrv.CorpOriginID = clhs.ID()
	tcrv.CorpTargetID = crhs.ID()
	tcrv.CityOriginID = lhs.ID()
	tcrv.CityTargetID = rhs.ID()
	tcrv.MapID = grd.ID()
	tcrv.Exported.ItemType = []string{"Iron"}
	tcrv.Exported.Quality.Min = 5
	tcrv.Exported.Quality.Max = 50
	tcrv.Exported.Quantity.Min = 5
	tcrv.Exported.Quantity.Max = 10

	tcrv.Imported.ItemType = []string{"Wood"}
	tcrv.Imported.Quality.Min = 5
	tcrv.Imported.Quality.Max = 50
	tcrv.Imported.Quantity.Min = 5
	tcrv.Imported.Quantity.Max = 10

	tcrv.NextChange = tools.RoundTime(time.Now().UTC())
	tcrv.LastChange = tools.RoundTime(time.Now().UTC())
	tcrv.EndOfTerm = tools.AboutNow(30)

	tcrv.TravelingDistance = 1
	tcrv.TravelingSpeed = 1
	tcrv.LoadingDelay = 1

	// forcefully build it in the past ;)

	err := tcrv.Insert(dbh)
	if err != nil {

		log.Printf("Caravan %+v", tcrv)
		log.Fatalf("Err %s", err)
	}
	caravan_manager.GenerateHandler(tcrv)

	crv, _ := caravan_manager.GetCaravanHandler(tcrv.ID)

	//ensure both parties have got enough items to provide the contract.

	// note WE MUST BE QUICK, as grid_manager.Handler is timed to perform auto test soon ;)

	var it item.Item
	it.BasePrice = 15
	it.Type = []string{"Iron"}
	it.Name = "Iron Ingot"
	it.Quality = 15
	it.Quantity = 100 // Should be enough for now.

	lhs.Call(func(city *city.City) {
		city.Storage.Add(it)
	})

	var it2 item.Item
	it2.BasePrice = 15
	it2.Type = []string{"Wood"}
	it2.Name = "Pine Logs"
	it2.Quality = 15
	it2.Quantity = 100

	rhs.Call(func(city *city.City) {
		city.Storage.Add(it2)
	})

	// We're now ready to gooo :)

	// Identifying all participants:

	log.Printf("############ City Origin: %d %d", crv.Get().CityOriginID, lhs.ID())
	log.Printf("############ Corp Origin: %d %d", crv.Get().CorpOriginID, clhs.ID())
	log.Printf("############ City Target: %d %d", crv.Get().CityTargetID, rhs.ID())
	log.Printf("############ Corp Target: %d %d", crv.Get().CorpTargetID, crhs.ID())

	// First step first, Recipient accept the contract !
	crv.Call(func(caravan *caravan.Caravan) {
		dbh = db.NewTest()
		defer dbh.Close()
		err := caravan.Accept(dbh, crhs.ID())
		if err != nil {
			log.Printf("Caravan: Failed to accept %s", err)
		}
	})

	log.Printf("######### Finished generation of Base State: Caravan %+v %s", crv.Get(), crv.Get().FullStringState())

	// forcefully mark caravan to be ready to arrive ;)
	crv.Call(func(crvn *caravan.Caravan) {
		now := tools.RoundNow()
		log.Printf("Crv: Moved time to %s", now.Format(time.RFC3339))

		// forcefully mark caravan to be ready to go ;)
		if crvn.State != caravan.CRVWaitingOriginLoad {
			log.Printf("Caravan state %+v", crv.Get())
			t.Errorf("Caravan wasn't in state Waiting origin load")
			return
		}

		crvn.NextChange = now

		hlhs, _ := city_manager.GetCityHandler(crvn.CityOriginID)
		hrhs, _ := city_manager.GetCityHandler(crvn.CityTargetID)
		clhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpOriginID)
		crhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpTargetID)
		crvn.PerformNextStep(hlhs, hrhs, clhs, crhs, now)

	})

	// ensure caravan current state is set to 5

	crv.Call(func(crvn *caravan.Caravan) {

		if crvn.State != caravan.CRVTravelingToTarget {
			t.Errorf("Caravan wasn't in state traveling to target")
			return
		}
		// forcefully mark caravan to be ready to arrive ;)
		now := tools.RoundNow()
		log.Printf("Crv: Moved time to %s", now.Format(time.RFC3339))

		crvn.NextChange = now
		hlhs, _ := city_manager.GetCityHandler(crvn.CityOriginID)
		hrhs, _ := city_manager.GetCityHandler(crvn.CityTargetID)
		clhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpOriginID)
		crhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpTargetID)
		crvn.PerformNextStep(hlhs, hrhs, clhs, crhs, now)
	})

	// ensure caravan current state is set to 7

	crv.Call(func(crvn *caravan.Caravan) {
		if crvn.State != caravan.CRVWaitingTargetLoad {
			t.Errorf("Caravan wasn't in state waiting for target load")
			return
		}
		// forcefully mark caravan to be ready to arrive ;)
		now := tools.RoundNow()
		crvn.NextChange = now
		log.Printf("Crv: Moved time to %s", now.Format(time.RFC3339))

		hlhs, _ := city_manager.GetCityHandler(crvn.CityOriginID)
		hrhs, _ := city_manager.GetCityHandler(crvn.CityTargetID)
		clhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpOriginID)
		crhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpTargetID)
		crvn.PerformNextStep(hlhs, hrhs, clhs, crhs, now)
	})

	crv.Call(func(crvn *caravan.Caravan) {
		// ensure caravan current state is set to 8
		if crvn.State != caravan.CRVTravelingToOrigin {
			t.Errorf("Caravan wasn't in state traveling to origin")
			return
		}
		// forcefully mark caravan to be ready to arrive ;)
		now := tools.RoundNow()
		crvn.NextChange = now
		hlhs, _ := city_manager.GetCityHandler(crvn.CityOriginID)
		hrhs, _ := city_manager.GetCityHandler(crvn.CityTargetID)
		clhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpOriginID)
		crhs, _ := corporation_manager.GetCorporationHandler(crvn.CorpTargetID)
		crvn.PerformNextStep(hlhs, hrhs, clhs, crhs, now)
	})

	// Must be back to square 4 ;)
	// ensure caravan current state is set to 4

	if crv.Get().State != caravan.CRVWaitingOriginLoad {
		t.Errorf("Caravan wasn't in state waiting origin load")
		return
	}

}
