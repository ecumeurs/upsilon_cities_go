package grid_controller

import (
	"encoding/json"
	"net/http"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

// Index GET: /map
func Index(w http.ResponseWriter, req *http.Request) {
	// handler := db.New()
	// defer handler.Close()

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
	// gard := context.Get(req, "map").(*grid.Grid)
	grd := grid.New()
	data := prepareGrid(grd)
	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, "map\\show", data)
	}
}
