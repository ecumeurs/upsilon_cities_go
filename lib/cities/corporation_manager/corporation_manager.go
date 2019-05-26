package corporation_manager

import (
	"errors"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/misc/actor"
)

//Handler own the carava, they're to be called upon to provide access to the carava
//It's also the one responsible for handling corporation lifecyle ...
type Handler struct {
	*actor.Actor
	corp *corporation.Corporation
}

//Manager keeps track of grid handlers out there.
//Should also have a TTL running on them ... but well ;)
type Manager struct {
	*actor.Actor
	handlers  map[int]*Handler
	ByCityIDs map[int]int
	ByMapID   map[int][]int
	ender     chan<- actor.End
}

var manager Manager

//Get access to read only version of the corpo ( a copy ) ... Still the store is still valid :'( but shouldn't be used.
func (h *Handler) Get() corporation.Corporation {
	return *h.corp
}

//InitManager initialize manager.
func InitManager() {
	manager.ender = make(chan actor.End)
	manager.Actor = actor.New(0, manager.ender)
	manager.handlers = make(map[int]*Handler)
	manager.ByCityIDs = make(map[int]int)
	manager.ByMapID = make(map[int][]int)
	manager.Start()
}

//GenerateHandler register a new handler for city.
func GenerateHandler(corp *corporation.Corporation) {

	cm := new(Handler)
	cm.corp = corp
	cm.Actor = actor.New(corp.ID, manager.ender)
	cm.Start()

	manager.Cast(func() {
		manager.handlers[corp.ID] = cm
		for _, v := range cm.Get().CitiesID {
			manager.ByCityIDs[v] = corp.ID
		}
		manager.ByMapID[cm.Get().MapID] = append(manager.ByMapID[cm.Get().MapID], corp.ID)
	})
}

//GetCorporationHandler Fetches corp from memory
func GetCorporationHandler(id int) (*Handler, error) {
	cm, found := manager.handlers[id]
	if found {
		return cm, nil
	}

	return nil, errors.New("unknown corp")
}

//GetCorporationHandlerByCityID Fetches grid from memory
func GetCorporationHandlerByCityID(cityID int) (res *Handler, err error) {

	cm, found := manager.ByCityIDs[cityID]
	if found {
		return GetCorporationHandler(cm)
	}
	return nil, errors.New("unknown city, no corp")
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
func (h *Handler) Cast(fn func(*corporation.Corporation)) {
	fn2 := func() {
		fn(h.corp)
	}
	h.Actor.Cast(fn2)
}

//Call send and wait for end of execution. Will provide access to protected grid
func (h *Handler) Call(fn func(*corporation.Corporation)) {
	fn2 := func() {
		fn(h.corp)
	}
	h.Actor.Call(fn2)
}
