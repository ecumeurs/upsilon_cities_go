//Package city_controller will allow manipulation of Cities and information gathering ...
package city_controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

type simpleNeighbourg struct {
	ID       int
	Location node.Point
	Name     string
}
type simpleCity struct {
	ID         int
	Location   node.Point
	Neighbours []simpleNeighbourg
	Name       string
}

func prepareSingleCity(cty *city.City) (res simpleCity) {
	res.ID = cty.ID
	res.Name = cty.Name
	res.Location = cty.Location
	for _, v := range cty.NeighboursID {
		callback := make(chan simpleNeighbourg)
		defer close(callback)
		cm, _ := city_manager.GetCityHandler(v)
		cm.Cast(func(c *city.City) {
			var sn simpleNeighbourg
			sn.ID = c.ID
			sn.Location = c.Location
			sn.Name = c.Name
			callback <- sn
		})
		res.Neighbours = append(res.Neighbours, <-callback)
	}
	return
}

func prepareCity(cities map[int]*city.City, city *city.City) (res simpleCity) {
	res.ID = city.ID
	res.Name = city.Name
	res.Location = city.Location
	for _, v := range city.NeighboursID {
		var sn simpleNeighbourg
		sn.ID = v
		sn.Location = cities[v].Location
		sn.Name = cities[v].Name
		res.Neighbours = append(res.Neighbours, sn)
	}
	return
}

func prepareCities(cities map[int]*city.City) (res []simpleCity) {
	for _, v := range cities {
		res = append(res, prepareCity(cities, v))
	}
	return
}

// Index GET: /map/:map_id/cities
// Provide basic informations on all cities available on given map.
// NOTE all city templates will be displayed without any layout ( meant to be serviced by API and replace HTML by json )
func Index(w http.ResponseWriter, req *http.Request) {
	id, err := tools.GetInt(req, "map_id")
	if err != nil {
		tools.Fail(w, req, "Provided map_id isn't an integer", "/map")
		return
	}

	handler := db.New()
	defer handler.Close()

	gm, err := grid_manager.GetGridHandler(id)

	if err != nil {
		tools.Fail(w, req, "Unknown map id", fmt.Sprintf("/map"))
		return
	}

	callback := make(chan []simpleCity)
	defer close(callback)
	gm.Cast(func(grd *grid.Grid) {
		callback <- prepareCities(grd.Cities)
	})

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(<-callback)
	} else {

		tools.GenerateAPIError(w, "Accessible only through API")
	}
}

// Show GET: /map/:map_id/city/:city_id
func Show(w http.ResponseWriter, req *http.Request) {
	city_id, err := tools.GetInt(req, "city_id")
	map_id, err := tools.GetInt(req, "map_id")

	cm, err := city_manager.GetCityHandler(city_id)
	if err != nil {
		tools.Fail(w, req, "Unknown city id", fmt.Sprintf("/map/%d", map_id))
		return
	}

	cmcallback := make(chan simpleCity)
	defer close(cmcallback)

	cm.Cast(func(city *city.City) {
		cmcallback <- prepareSingleCity(city)
	})

	data := <-cmcallback

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, "city\\show", data)
	}
}
