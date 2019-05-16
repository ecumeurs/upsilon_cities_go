package grid_manager

import (
	"errors"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/actor"
)

//Handler own the grid, they're to be called upon to provide access to the grid
type Handler struct {
	*actor.Actor
	grid *grid.Grid
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

	grd = new(Handler)
	grd.grid = gd
	grd.Actor = actor.New(id, manager.ender)
	grd.Start()

	manager.Cast(func() { manager.handlers[id] = grd })

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
