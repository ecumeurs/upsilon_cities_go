package grid_manager

import (
	"errors"
	"log"
	"time"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	grid_evolution "upsilon_cities_go/lib/cities/evolution/grid"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/actor"
)

//Handler own the grid, they're to be called upon to provide access to the grid
type Handler struct {
	*actor.Actor
	grid    *grid.Grid
	Ticker  *time.Ticker
	Deleted bool
}

//Manager keeps track of grid handlers out there.
//Should also have a TTL running on them ... but well ;)
type Manager struct {
	*actor.Actor
	handlers map[int]*Handler
	ender    chan<- actor.End
}

var manager Manager

//InitManager initialize manager.
func InitManager() {
	manager.ender = make(chan actor.End)
	manager.Actor = actor.New(0, manager.ender)
	manager.handlers = make(map[int]*Handler)
	manager.Start()
}

//Get access to a copy of grid.
func (g *Handler) Get() grid.Grid {
	return *g.grid
}

//GenerateGridHandler create a new grid handler and load related ressources.
func GenerateGridHandler(gd *grid.Grid) {

	grd := new(Handler)
	grd.grid = gd
	grd.Deleted = false
	grd.Actor = actor.New(gd.ID, manager.ender)
	grd.Ticker = time.NewTicker(tools.CycleLength * 10)
	grd.Loop = func() {
		for {
			select {
			case <-grd.Ticker.C:
				if grd.Deleted {
					log.Fatalf("GridMgr: Should have been deleted but wasn't ... %d", grd.ID())
					return
				}
				grid_evolution.UpdateRegion(grd.grid)
			case f := <-grd.Actionc:
				if grd.Deleted {
					log.Fatalf("GridMgr: Should have been deleted but wasn't ... %d", grd.ID())
					return
				}
				f()
			case <-grd.Quitc:
				return
			}
		}
	}
	grd.Start()

	dbh := db.New()
	defer dbh.Close()

	gd.Cities, _ = city.ByMap(dbh, grd.ID())

	for _, v := range gd.Cities {
		city_manager.GenerateHandler(v)
		log.Printf("Grid: Created City Handler %d %s", v.ID, v.Name)
	}

	caravans, _ := caravan.ByMapID(dbh, gd.ID)

	for _, v := range caravans {
		caravan_manager.GenerateHandler(v)
		log.Printf("Grid: Created Caravan Handler %d", v.ID)
	}

	corps, _ := corporation.ByMapID(dbh, gd.ID)

	for _, v := range corps {
		corporation_manager.GenerateHandler(v)
		log.Printf("Grid: Created Corp Handler %d %s", v.ID, v.Name)
	}

	// ensure evolution gets kicked in.
	grid_evolution.LoadEvolution(grd.grid)

	// might as well add ticker in place ;)

	manager.Cast(func() { manager.handlers[gd.ID] = grd })

}

//GetGridHandler Fetches grid from memory
func GetGridHandler(id int) (*Handler, error) {
	grd, found := manager.handlers[id]
	if found {
		return grd, nil
	}

	dbh := db.New()
	defer dbh.Close()

	gd, err := grid.ByID(dbh, id)

	if err != nil {
		return nil, err
	}

	GenerateGridHandler(gd)

	grd, _ = manager.handlers[id]
	return grd, nil
}

//DropGridHandler from memory
func DropGridHandler(id int) error {
	grid, found := manager.handlers[id]
	if !found {
		return errors.New("Unable to drop non existant Grid")
	}

	manager.Cast(func() {
		delete(manager.handlers, id)
		grid.Deleted = true
		grid.Ticker.Stop()
		grid.Stop()
	})

	return nil
}

//Cast send and forget. Will provide access to protected grid.
// If you want a reply, dont forget to provide your function a chan
// If you do so, DONT FORGET TO call defer close(<your chan>)
func (a *Handler) Cast(fn func(*grid.Grid)) {
	fn2 := func() {
		fn(a.grid)
	}

	a.Actor.Cast(fn2)
}

//Call send and wait for end of execution. Will provide access to protected grid
func (a *Handler) Call(fn func(*grid.Grid)) {
	fn2 := func() {
		fn(a.grid)
	}

	a.Actor.Call(fn2)
}
