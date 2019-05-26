package grid

import (
	"log"
	"time"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

//LoadEvolution restore state of the grid
// will seek out every evolving parameter of the grid and keep track of the next important date.
// That is (for the moment) next caravan reaching a destination.
func (grid *Grid) LoadEvolution() {
	// seek access to all cities and caravans.
	// We're in a read only environment ... so mayhaps we could do this quick and nicely.

	// grid has a dumbass access to cities ... it's used to seed city manager ... whatever.
	grid.SeekNextCaravan()
}

//SeekNextCaravan seek next caravan cycle date. As it will impact a city.
func (grid *Grid) SeekNextCaravan() {

	nextUpdate := tools.AddCycles(tools.RoundNow(), 1000)
	nextCrv := 0
	// hopefully will not abuse of this state ...
	for k := range grid.Cities {
		// we only need id ;)
		chs, err := caravan_manager.GetCaravanHandlerByCityID(k)
		if err != nil {
			log.Printf("Grid: Evolution loading ... failed to access to city %d caravans %s", k, err)
			continue
		}

		for _, v := range chs {
			// same, shouldn't abuse of this one:
			crv := v.Get()
			//
			if crv.IsProducing() && nextUpdate.After(crv.NextChange) {
				nextUpdate = crv.NextChange
				nextCrv = crv.ID
			}
		}
	}

	grid.Evolution.NextCaravan = nextUpdate
	grid.Evolution.NextCaravanID = nextCrv
}

//UpdateRegion performed from within grid thread.
// Will update the whole region up to now ;)
func (grid *Grid) UpdateRegion() {
	rnow := tools.RoundNow()

	if rnow.Equal(grid.LastUpdate) {
		// nothing to do anyway
		log.Printf("Grid: No last update is too recent.")
		return
	}

	// check if a caravan will be finished before now, and so long now isn't reached continue on.

	nextStop := tools.MinTime(rnow, grid.Evolution.NextCaravan)
	for nextStop.Before(rnow) {

		crv, err := caravan_manager.GetCaravanHandler(grid.Evolution.NextCaravanID)
		if err != nil {
			log.Printf("Grid: Unable to find caravan to update ..")
		}
		for k := range grid.Cities {
			cm, _ := city_manager.GetCityHandler(k)
			cm.Cast(func(city *city.City) {
				city.CheckActivity(nextStop)
			})
		}

		// this IS highly dangerous ;)
		crv.Cast(func(caravan *caravan.Caravan) {
			hlhs, _ := city_manager.GetCityHandler(caravan.CityOriginID)
			hrhs, _ := city_manager.GetCityHandler(caravan.CityTargetID)
			dbh := db.New()
			defer dbh.Close()
			caravan.PerformNextStep(dbh, hlhs, hrhs, nextStop)
		})

		grid.SeekNextCaravan()
		nextStop = tools.MinTime(rnow, grid.Evolution.NextCaravan)
	}

	for k := range grid.Cities {
		cm, _ := city_manager.GetCityHandler(k)
		dbh := db.New()
		defer dbh.Close()
		cm.Call(func(city *city.City) {
			city.CheckCityOwnership(dbh)
			city.Update(dbh)
		})
	}

	grid.LastUpdate = rnow
	grid.SeekNextCaravan()
	log.Printf("Grid: Update done, next caravan: %s", grid.Evolution.NextCaravan.Format(time.RFC3339))
}

//RegionUpdateNeeded tell whether the whole region need to get updated for this city to get updated...
func (grid *Grid) RegionUpdateNeeded(cityID int) bool {
	cm, err := city_manager.GetCityHandler(cityID)
	if err != nil {
		log.Printf("Grid: Unable to check state of city %d", cityID)
		return false
	}

	rnow := tools.RoundNow()

	if rnow.Equal(grid.LastUpdate) {
		// nothing to do anyway
		log.Printf("Grid: No last update is too recent.")
		return false
	}

	if rnow.Before(grid.Evolution.NextCaravan) {
		return false
	}

	if cm.Get().NextUpdate.After(grid.Evolution.NextCaravan) {
		return false
	}

	// okay so next update for this city comes after next caravan so that might be okay ...
	return true // simpler that way ...
}
