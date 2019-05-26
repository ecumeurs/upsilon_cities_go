package city_manager

import (
	"errors"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/misc/actor"
)

//Handler own the grid, they're to be called upon to provide access to the grid
type Handler struct {
	*actor.Actor
	city *city.City
}

//Manager keeps track of grid handlers out there.
//Should also have a TTL running on them ... but well ;)
type Manager struct {
	*actor.Actor
	handlers map[int]*Handler
	ByMap    map[int][]int
	ender    chan<- actor.End
}

var manager Manager

//Get access to read only version of the caravan ( a copy ) ... Still the store is still valid :'( but shouldn't be used.
func (h *Handler) Get() city.City {
	return *h.city
}

//InitManager initialize manager.
func InitManager() {
	manager.ender = make(chan actor.End)
	manager.Actor = actor.New(0, manager.ender)
	manager.handlers = make(map[int]*Handler)
	manager.ByMap = make(map[int][]int)
	manager.Start()
}

//GenerateHandler register a new handler for city.
func GenerateHandler(city *city.City) {

	cty := new(Handler)
	cty.city = city
	cty.Actor = actor.New(city.ID, manager.ender)
	cty.Start()

	manager.Cast(func() {
		manager.handlers[city.ID] = cty
		manager.ByMap[city.MapID] = append(manager.ByMap[city.MapID], city.ID)
	})

}

//GetCityHandler Fetches grid from memory
func GetCityHandler(id int) (*Handler, error) {
	cty, found := manager.handlers[id]
	if found {
		return cty, nil
	}

	return nil, errors.New("city hasn't been loaded")
}

//GetCityHandlerByMap Fetches grid from memory
func GetCityHandlerByMap(id int) (res []*Handler, err error) {
	cty, found := manager.ByMap[id]
	if found {
		for _, v := range cty {
			res = append(res, manager.handlers[v])
		}
	}
	return
}

//DropCityHandler from memory
func DropCityHandler(id int) error {
	city, found := manager.handlers[id]
	if !found {
		return errors.New("Unable to drop non existant Grid")
	}

	manager.Cast(func() {
		delete(manager.handlers, id)
		city.Stop()
	})

	return nil
}

//Cast send and forget. Will provide access to protected grid.
// If you want a reply, dont forget to provide your function a chan
// If you do so, DONT FORGET TO call defer close(<your chan>)
func (a *Handler) Cast(fn func(*city.City)) {
	fn2 := func() {
		fn(a.city)
	}
	a.Actor.Cast(fn2)
}

//Call send and wait for end of execution. Will provide access to protected grid
func (a *Handler) Call(fn func(*city.City)) {
	fn2 := func() {
		fn(a.city)
	}
	a.Actor.Call(fn2)
}
