package caravan_manager

import (
	"errors"
	"upsilon_cities_go/lib/cities/caravan"
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
	ByCorpIDs map[int][]int
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
	manager.ByCityIDs = make(map[int][]int)
	manager.ByCorpIDs = make(map[int][]int)
	manager.ByMapID = make(map[int][]int)
	manager.Start()
}

//GenerateHandler register a new handler for city.
func GenerateHandler(caravan *caravan.Caravan) {

	cm := new(Handler)
	cm.caravan = caravan
	cm.Actor = actor.New(caravan.ID, manager.ender)
	cm.Start()

	manager.Cast(func() {
		manager.handlers[caravan.ID] = cm
		manager.ByCityIDs[caravan.CityOriginID] = append(manager.ByCityIDs[caravan.CityOriginID], caravan.ID)
		manager.ByCityIDs[caravan.CityTargetID] = append(manager.ByCityIDs[caravan.CityTargetID], caravan.ID)
		manager.ByCorpIDs[caravan.CorpOriginID] = append(manager.ByCorpIDs[caravan.CorpOriginID], caravan.ID)
		manager.ByCorpIDs[caravan.CorpTargetID] = append(manager.ByCorpIDs[caravan.CorpTargetID], caravan.ID)

		manager.ByMapID[caravan.MapID] = append(manager.ByMapID[caravan.MapID], caravan.ID)
	})

}

//GetCaravanHandler Fetches grid from memory
func GetCaravanHandler(id int) (*Handler, error) {
	cm, found := manager.handlers[id]
	if found {
		return cm, nil
	}

	return nil, errors.New("unknown caravan, reload webpage")
}

//GetCaravanHandlerByCorpID Fetches grid from memory
func GetCaravanHandlerByCorpID(corpID int) (res []*Handler, err error) {

	cm, found := manager.ByCorpIDs[corpID]
	if found {
		for _, v := range cm {
			res = append(res, manager.handlers[v])
		}
		return res, nil
	}

	return
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

	return
}

//DropCaravanHandler from memory
func DropCaravanHandler(id int) error {
	cm, found := manager.handlers[id]
	if !found {
		return errors.New("Unable to drop non existant Caravan")
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
func (h *Handler) Cast(fn func(*caravan.Caravan)) {
	fn2 := func() {
		fn(h.caravan)
	}
	h.Actor.Cast(fn2)
}

//Call send and wait for end of execution. Will provide access to protected grid
func (h *Handler) Call(fn func(*caravan.Caravan)) {
	fn2 := func() {
		fn(h.caravan)
	}
	h.Actor.Call(fn2)
}
