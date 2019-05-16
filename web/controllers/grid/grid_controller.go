package grid_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

// Index GET: /map
func Index(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}
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

// Show GET: /map/:id also: stores current_corp_id in session.
func Show(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}

	id, err := tools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	uid, err := tools.CurrentUserID(req)
	dbh := db.New()
	defer dbh.Close()
	corp, err := corporation.ByMapIDByUserID(dbh, id, uid)

	if err != nil {
		if tools.IsAPI(req) {
			tools.GenerateAPIOk(w)
			json.NewEncoder(w).Encode(fmt.Sprintf("Need to select a corporation, call /api/map/%d/select_corporation", id))
		} else {
			tools.Redirect(w, req, fmt.Sprintf("/map/%d/select_corporation", id))
		}
		return
	}

	tools.GetSession(req).Values["current_corp_id"] = corp.ID

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

type shortCorporation struct {
	ID   int
	Name string
}

//ShowSelectableCorporation GET /map/:map_id/select_corporation allow one use to select a claimable corporation.
func ShowSelectableCorporation(w http.ResponseWriter, req *http.Request) {
	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}

	id, err := tools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	uid, err := tools.CurrentUserID(req)
	dbh := db.New()
	defer dbh.Close()
	_, err = corporation.ByMapIDByUserID(dbh, id, uid)

	if err == nil {
		// has already selected a corporation for this map ...

		if tools.IsAPI(req) {
			tools.GenerateAPIOkAndSend(w)
		} else {
			tools.Redirect(w, req, fmt.Sprintf("/map/%d", id))
		}
		return
	}

	corps, err := corporation.ByMapIDClaimable(dbh, id)

	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Unable to find claimable corporations.", "/map")
		return
	}

	if len(corps) == 0 {
		// failed to convert id to int ...
		tools.Fail(w, req, "No corporations left to claim.", "/map")
		return
	}

	// create short corps ;)

	data := make([]shortCorporation, 0)

	for _, v := range corps {
		d := shortCorporation{v.ID, v.Name}
		data = append(data, d)
	}

	var res struct {
		Data  []shortCorporation
		MapID int
	}

	res.Data = data
	res.MapID = id

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(res)
	} else {
		templates.RenderTemplate(w, req, "map\\select_corp", res)
	}
}

//SelectCorporation POST /map/:map_id/select_corporation allow one use to select a claimable corporation.
func SelectCorporation(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}
	id, err := tools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	uid, err := tools.CurrentUserID(req)
	dbh := db.New()
	defer dbh.Close()
	_, err = corporation.ByMapIDByUserID(dbh, id, uid)

	if err == nil {
		// has already selected a corporation for this map ...

		if tools.IsAPI(req) {
			tools.GenerateAPIOkAndSend(w)
		} else {
			tools.Redirect(w, req, fmt.Sprintf("/map/%d", id))
		}
		return
	}

	req.ParseForm()
	f := req.Form
	corpID, err := strconv.Atoi(f.Get("corporation"))
	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Unable to find read corporation", "/map")
		return
	}

	corp, err := corporation.ByID(dbh, corpID)

	if err != nil {

		// failed to convert id to int ...
		tools.Fail(w, req, "Unable to find requested corporation", "/map")
		return
	}

	usr, err := tools.CurrentUser(req)

	err = corporation.Claim(dbh, usr, corp)

	if err != nil {
		// failed to convert id to int ...
		tools.Fail(w, req, "Unable to claim corporation", "/map")
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.Redirect(w, req, fmt.Sprintf("/map/%d", id))
	}
}

// Create POST: /map
func Create(w http.ResponseWriter, req *http.Request) {

	if !tools.IsAdmin(req) {
		tools.Fail(w, req, "must be admin", "/")
	}

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

	if !tools.IsAdmin(req) {
		tools.Fail(w, req, "must be admin", "/")
	}

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
