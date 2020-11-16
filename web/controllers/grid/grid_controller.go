package grid_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/grid_manager"
	"upsilon_cities_go/lib/cities/map/map_generator/region"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/user"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/webtools"
)

// Index GET: /map
func Index(w http.ResponseWriter, req *http.Request) {

	if !webtools.IsLogged(req) {
		webtools.Fail(w, req, "must be logged in", "/")
		return
	}

	uid, err := webtools.CurrentUserID(req)
	dbh := db.New()
	defer dbh.Close()

	grids, err := grid.AllShortened(dbh)
	if err != nil {
		webtools.Fail(w, req, "Failed to get all maps ...", "/")
		return
	}

	var dataList []indexGrid
	for _, localgrid := range grids {
		isUserOnMap, _ := user.IsUserOnMap(dbh, uid, localgrid.ID)
		if isUserOnMap {

			corp, err := corporation.ByMapIDByUserID(dbh, localgrid.ID, uid)
			// grid should be loaded first ... some stuff should be kept updated ;)
			if err != nil {
				// failed to find corporation.
				webtools.Fail(w, req, "An Error as occured.", "/map")
				return
			}
			var data indexGrid
			data.Name = localgrid.Name
			data.ID = localgrid.ID
			data.UserCorp.ID = corp.ID
			data.UserCorp.Name = corp.Name
			data.UserCorp.Fame = 0
			data.UserCorp.Credits = corp.Credits
			crvs, _ := caravan_manager.GetCaravaRequiringAction(corp.ID)
			data.UserCorp.CrvWaiting = len(crvs)

			dataList = append(dataList, data)

		} else {
			var data indexGrid
			data.Name = localgrid.Name
			data.ID = localgrid.ID
			dataList = append(dataList, data)
		}
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(grids)
	} else {
		webtools.GetSession(req).Values["current_corp_id"] = 0
		templates.RenderTemplate(w, req, "map/index", dataList)
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

type simpleCorp struct {
	ID         int
	Name       string
	Credits    int
	Fame       int
	CrvWaiting int
}

type gameInfo struct {
	WebGrid  webGrid
	UserCorp simpleCorp
}

type indexGrid struct {
	UserCorp simpleCorp
	Name     string
	ID       int
}

// Show GET: /map/:id also: stores current_corp_id in session.
func Show(w http.ResponseWriter, req *http.Request) {

	if !webtools.IsLogged(req) {
		webtools.Fail(w, req, "must be logged in", "/")
		return
	}

	id, err := webtools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		webtools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	// grid should be loaded first ... some stuff should be kept updated ;)
	grd, err := grid_manager.GetGridHandler(id)
	if err != nil {
		// failed to find requested map.
		webtools.Fail(w, req, "Unknown map id", "/map")
		return
	}

	uid, err := webtools.CurrentUserID(req)
	dbh := db.New()
	defer dbh.Close()
	corp, err := corporation.ByMapIDByUserID(dbh, id, uid)

	if err != nil {
		if webtools.IsAPI(req) {
			webtools.GenerateAPIOk(w)
			json.NewEncoder(w).Encode(fmt.Sprintf("Need to select a corporation, call /api/map/%d/select_corporation", id))
		} else {
			webtools.Redirect(w, req, fmt.Sprintf("/map/%d/select_corporation", id))
		}
		return
	}

	webtools.GetSession(req).Values["current_corp_id"] = corp.ID

	callback := make(chan gameInfo)
	defer close(callback)
	grd.Cast(func(grid *grid.Grid) {
		var grd webGrid
		grd.Nodes = prepareGrid(grid)
		grd.Name = grid.Name
		var ginf gameInfo
		ginf.WebGrid = grd
		ginf.UserCorp.ID = corp.ID
		ginf.UserCorp.Name = corp.Name
		ginf.UserCorp.Credits = corp.Credits
		crvs, _ := caravan_manager.GetCaravaRequiringAction(corp.ID)
		ginf.UserCorp.CrvWaiting = len(crvs)
		callback <- ginf
	})

	var data gameInfo
	data = <-callback
	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "map/show", data)
	}
}

type shortCorporation struct {
	ID   int
	Name string
}

//ShowSelectableCorporation GET /map/:map_id/select_corporation allow one use to select a claimable corporation.
func ShowSelectableCorporation(w http.ResponseWriter, req *http.Request) {
	if !webtools.IsLogged(req) {
		webtools.Fail(w, req, "must be logged in", "/")
		return
	}

	id, err := webtools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		webtools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	uid, err := webtools.CurrentUserID(req)
	dbh := db.New()
	defer dbh.Close()
	_, err = corporation.ByMapIDByUserID(dbh, id, uid)

	if err == nil {
		// has already selected a corporation for this map ...

		if webtools.IsAPI(req) {
			webtools.GenerateAPIOkAndSend(w)
		} else {
			webtools.Redirect(w, req, fmt.Sprintf("/map/%d", id))
		}
		return
	}

	corps, err := corporation.ByMapIDClaimable(dbh, id)

	if err != nil {
		// failed to convert id to int ...
		webtools.Fail(w, req, "Unable to find claimable corporations.", "/map")
		return
	}

	if len(corps) == 0 {
		// failed to convert id to int ...
		webtools.Fail(w, req, "No corporations left to claim.", "/map")
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

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(res)
	} else {
		templates.RenderTemplate(w, req, "map/select_corp", res)
	}
}

//SelectCorporation POST /map/:map_id/select_corporation allow one use to select a claimable corporation.
func SelectCorporation(w http.ResponseWriter, req *http.Request) {

	if !webtools.IsLogged(req) {
		webtools.Fail(w, req, "must be logged in", "/")
		return
	}
	id, err := webtools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		webtools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	uid, err := webtools.CurrentUserID(req)
	dbh := db.New()
	defer dbh.Close()
	_, err = corporation.ByMapIDByUserID(dbh, id, uid)

	if err == nil {
		// has already selected a corporation for this map ...

		if webtools.IsAPI(req) {
			webtools.GenerateAPIOkAndSend(w)
		} else {
			webtools.Redirect(w, req, fmt.Sprintf("/map/%d", id))
		}
		return
	}

	req.ParseForm()
	f := req.Form
	corpID, err := strconv.Atoi(f.Get("corporation"))
	if err != nil {
		// failed to convert id to int ...
		webtools.Fail(w, req, "Unable to find read corporation", "/map")
		return
	}

	cm, err := corporation_manager.GetCorporationHandler(corpID)

	if err != nil {

		// failed to convert id to int ...
		webtools.Fail(w, req, "Unable to find requested corporation", "/map")
		return
	}

	usr, err := webtools.CurrentUser(req)

	cb := make(chan error)
	defer close(cb)

	cm.Cast(func(corp *corporation.Corporation) {
		err = corporation.Claim(dbh, usr, corp)
		cb <- err
	})

	err = <-cb
	if err != nil {
		// failed to convert id to int ...
		webtools.Fail(w, req, "Unable to claim corporation", "/map")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, fmt.Sprintf("/map/%d", id))
	}
}

// Create POST: /map
func Create(w http.ResponseWriter, req *http.Request) {

	if !webtools.IsAdmin(req) {
		webtools.Fail(w, req, "must be admin", "/")
		return
	}

	var grd *grid.Grid
	handler := db.New()
	defer handler.Close()
	reg, err := region.Generate("Elvenwood")
	if err != nil {
		webtools.Fail(w, req, "Unable to create an Elvenwood map", "/")
		return
	}

	grd, err = reg.Generate(handler)
	if err != nil {
		webtools.Fail(w, req, "Unable to generate an Elvenwood map", "/")
		return
	}

	grid.Store(handler, grd)
	log.Printf("GC: Store map: \n%s", grd.String())

	grid_manager.GenerateGridHandler(grd)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(grd)
	} else {
		webtools.Redirect(w, req, fmt.Sprintf("/map/%d", grd.ID))
	}
}

//Destroy DELETE: /map/:id
func Destroy(w http.ResponseWriter, req *http.Request) {

	if !webtools.IsAdmin(req) {
		webtools.Fail(w, req, "must be admin", "/")
		return
	}

	handler := db.New()
	defer handler.Close()

	id, err := webtools.GetInt(req, "map_id")
	if err != nil {
		// failed to convert id to int ...
		webtools.Fail(w, req, "Invalid map id format", "/map")
		return
	}

	log.Printf("GridCtrl: About to delete map %d", id)

	grid_manager.DropGridHandler(id)
	grid.DropByID(handler, id)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("success")
	} else {
		webtools.Redirect(w, req, "/map")
	}
}
