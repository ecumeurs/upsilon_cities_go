//Package city_controller will allow manipulation of Cities and information gathering ...
package city_controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/grid"
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

func prepareCity(city *city.City) (res simpleCity) {
	res.ID = city.ID
	res.Name = city.Name
	res.Location = city.Location
	for _, v := range city.Neighbours {
		var sn simpleNeighbourg
		sn.ID = v.ID
		sn.Location = v.Location
		sn.Name = v.Name
		res.Neighbours = append(res.Neighbours, sn)
	}
	return
}

func prepareCities(cities map[int]*city.City) (res []simpleCity) {
	for _, v := range cities {
		res = append(res, prepareCity(v))
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

	cities, err := city.ByMap(handler, id)

	if err != nil {
		tools.Fail(w, req, "Failed to fetch maps cities ...", fmt.Sprintf("/map/%d", id))
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(prepareCities(cities))
	} else {

		tools.GenerateAPIError(w, "Accessible only through API")
	}
}

// Show GET: /map/:map_id/city/:city_id
func Show(w http.ResponseWriter, req *http.Request) {
	handler := db.New()
	defer handler.Close()

	city_id, err := tools.GetInt(req, "city_id")
	map_id, err := tools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Invalid city_id format", fmt.Sprintf("/map/%d", map_id))
		return
	}

	grd, err := grid.ByID(handler, map_id)
	if err != nil {
		// failed to find requested map.
		tools.Fail(w, req, "Unknown map id", "/map")
		return
	}

	city, found := grd.Cities[city_id]
	if !found {
		// failed to find requested map.
		tools.Fail(w, req, "Unknown city id", fmt.Sprintf("/map/%d", map_id))
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(prepareCity(city))
	} else {
		templates.RenderTemplate(w, "city\\show", prepareCity(city))
	}
}
