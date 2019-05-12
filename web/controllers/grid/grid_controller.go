package grid_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

// Index GET: /map
func Index(w http.ResponseWriter, req *http.Request) {

	handler := db.New()
	defer handler.Close()

	grids, err := grid.AllShortened(handler)
	if err != nil {
		tools.Fail(w, req, "Failed to load all maps ...", "/")
		return
	}
	// data := gardens.AllIds(handler)

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(grids)
	} else {
		templates.RenderTemplate(w, req, "map\\index", grids)
	}

}

type displayNode struct {
	Node       node.Node
	City       city.City
	Neighbours []node.Point
}

func prepareGrid(grd *grid.Grid) (res [][]displayNode) {

	var tmpRes []displayNode
	for _, nd := range grd.Nodes {
		var tmp displayNode
		tmp.Node = nd
		testCity := grd.GetCityByLocation(nd.Location)
		if testCity != nil {
			tmp.City = *testCity
			for _, v := range testCity.NeighboursID {
				tmp.Neighbours = append(tmp.Neighbours, grd.Cities[v].Location)
			}
		}

		tmpRes = append(tmpRes, tmp)
		if len(tmpRes) == grd.Size {
			res = append(res, tmpRes)
			tmpRes = make([]displayNode, 0, grd.Size)
		}
	}

	return
}

type webGrid struct {
	Nodes [][]displayNode
	Name  string
}

// Show GET: /map/:id
func Show(w http.ResponseWriter, req *http.Request) {
	id, err := tools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	grd, err := grid_manager.GetGridHandler(id)
	if err != nil {
		// failed to find requested map.
		tools.Fail(w, req, "Unknown map id", "/map")
		return
	}

	callback := make(chan webGrid)
	defer close(callback)
	grd.Cast(func(grid *grid.Grid) {
		var grd webGrid
		grd.Nodes = prepareGrid(grid)
		grd.Name = grid.Name
		callback <- grd
	})

	var data webGrid
	data = <-callback
	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "map\\show", data)
	}
}

// Create POST: /map
func Create(w http.ResponseWriter, req *http.Request) {
	var grd *grid.Grid
	handler := db.New()
	defer handler.Close()
	grd = grid.New(handler)

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(grd)
	} else {
		tools.Redirect(w, req, fmt.Sprintf("/map/%d", grd.ID))
	}
}

//Destroy DELETE: /map/:id
func Destroy(w http.ResponseWriter, req *http.Request) {

	handler := db.New()
	defer handler.Close()

	id, err := tools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	log.Printf("GridCtrl: About to delete map %d", id)

	grid.DropByID(handler, id)

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("success")
	} else {
		tools.Redirect(w, req, "/map")
	}
}
