package grid_controller

import (
	"encoding/json"
	"net/http"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

// Index GET: /map
func Index(w http.ResponseWriter, req *http.Request) {

	// data := gardens.AllIds(handler)

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(nil)
	} else {
		templates.RenderTemplate(w, "map\\index", "hello")
	}

}

type displayNode struct {
	Node node.Node
	City city.City
}

func prepareGrid(grd *grid.Grid) (res [][]displayNode) {

	var tmpRes []displayNode
	for _, nd := range grd.Nodes {
		var tmp displayNode
		tmp.Node = nd
		testCity := grd.GetCityByLocation(nd.Location)
		if testCity != nil {
			tmp.City = *testCity
		}
		tmpRes = append(tmpRes, tmp)
		if len(tmpRes) == grd.Size {
			res = append(res, tmpRes)
			tmpRes = make([]displayNode, 0, grd.Size)
		}
	}

	return
}

// Show GET: /map/:id
func Show(w http.ResponseWriter, req *http.Request) {
	random, found := tools.GetString(req, "map_id")
	if !found {
		// no id provided ??? impossible
		tools.Fail(w, req, "Unexpected request", "/map")
		return
	}

	var grd *grid.Grid
	handler := db.New()
	defer handler.Close()

	if random == "random" {
		// requesting random map
		grd = grid.New(handler)
	} else {
		id, err := tools.GetInt(req, "map_id")
		if err != nil {
			// failed to convert id to int ...
			tools.Fail(w, req, "Invalid map id format", "/map")
			return
		}

		grd, err = grid.ByID(handler, id)
		if err != nil {
			// failed to find requested map.
			tools.Fail(w, req, "Unknown map id", "/map")
			return
		}
	}

	data := prepareGrid(grd)
	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, "map\\show", data)
	}
}

// Create POST: /map/:id
func Create(w http.ResponseWriter, req *http.Request) {
}
