package caravan_manager

import (
	"errors"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/actor"
)

//Handler own the carava, they're to be called upon to provide access to the carava
//It's also the one responsible for handling caravan lifecyle ...
type Handler struct {
	*actor.Actor
	caravan *caravan.Caravan
}

//Manager keeps track of grid handlers out there.
//Should also have a TTL running on them ... but well ;)
type Manager struct {
	*actor.Actor
	handlers  map[int]*Handler
	ByCityIDs map[int][]int
	ByMapID   map[int][]int
	ender     chan<- actor.End
}

var manager Manager

//Get access to read only version of the caravan ( a copy ) ... Still the store is still valid :'( but shouldn't be used.
func (h *Handler) Get() caravan.Caravan {
	return *h.caravan
}

//InitManager initialize manager.
func InitManager() {
	manager.ender = make(chan actor.End)
	manager.Actor = actor.New(0, manager.ender)
	manager.handlers = make(map[int]*Handler)
	manager.Start()
}

//GetCaravanHandler Fetches grid from memory
func GetCaravanHandler(id int) (*Handler, error) {
	cm, found := manager.handlers[id]
	if found {
		return cm, nil
	}

	dbh := db.New()
	defer dbh.Close()

	caravan, err := caravan.ByID(dbh, id)

	if err != nil {
		return nil, err
	}

	cm = new(Handler)
	cm.caravan = caravan
	cm.Actor = actor.New(id, manager.ender)
	cm.Start()

	manager.Cast(func() { manager.handlers[id] = cm })

	return cm, nil
}

//GetCaravanHandlerByCityID Fetches grid from memory
func GetCaravanHandlerByCityID(cityID int) (res []*Handler, err error) {

	cm, found := manager.ByCityIDs[cityID]
	if found {
		for _, v := range cm {
			res = append(res, manager.handlers[v])
		}
		return res, nil
	}

	dbh := db.New()
	defer dbh.Close()

	caravans, err := caravan.ByCityID(dbh, cityID)

	if err != nil {
		return nil, err
	}

	for _, v := range caravans {
		ch := new(Handler)
		ch.caravan = v
		ch.Actor = actor.New(v.ID, manager.ender)
		ch.Start()

		manager.Cast(func() { manager.handlers[v.ID] = ch })
		res = append(res, ch)
	}

	return res, nil
}

//DropCaravanHandler from memory
func DropCaravanHandler(id int) error {
	cm, found := manager.handlers[id]
	if !found {
		return errors.New("Unable to drop non existant Grid")
	}

	manager.Cast(func() {
		delete(manager.handlers, id)
		cm.Stop()
	})

	return nil
}

//Cast send and forget. Will provide access to protected grid.
// If you want a reply, dont forget to provide your function a chan
// If you do so, DONT FORGET TO call defer close(<your chan>)
func (a *Handler) Cast(fn func(*caravan.Caravan)) {
	fn2 := func() {
		fn(a.caravan)
	}
	a.Actor.Cast(fn2)
}

//Call send and wait for end of execution. Will provide access to protected grid
func (a *Handler) Call(fn func(*caravan.Caravan)) {
	fn2 := func() {
		fn(a.caravan)
	}
	a.Actor.Call(fn2)
}
