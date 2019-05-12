//Package city_controller will allow manipulation of Cities and information gathering ...
package city_controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/node"
	lib_tools "upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

type simpleNeighbourg struct {
	ID       int
	Location node.Point
	Name     string
}

type simpleStorage struct {
	Count    int
	Capacity int
	Item     []item.Item
}

type simpleProducer struct {
	ProducerID   int
	ProductType  string
	ProductName  string
	Quality      lib_tools.IntRange
	Quantity     lib_tools.IntRange
	Active       bool
	EndTime      string
	Requirements string
}

type simpleCity struct {
	ID           int
	Location     node.Point
	Neighbours   []simpleNeighbourg
	NeighboursID []int
	Name         string
	Storage      simpleStorage
	Ressources   []simpleProducer
	Factories    []simpleProducer
	CorpoID      int
}

func prepareSingleCity(cm *city_manager.Handler) (res simpleCity) {
	callback := make(chan simpleCity)
	defer close(callback)

	cm.Cast(func(cty *city.City) {
		changed := cty.CheckActivity(time.Now().UTC())
		var rs simpleCity
		rs.ID = cty.ID
		rs.Name = cty.Name
		rs.Location = cty.Location
		rs.NeighboursID = cty.NeighboursID
		rs.Storage.Count = cty.Storage.Count()
		rs.Storage.Capacity = cty.Storage.Capacity
		rs.CorpoID = cty.CorporationID

		// To store the keys in slice in sorted order
		var keys []int64
		for k := range cty.Storage.Content {
			keys = append(keys, k)
		}

		lib_tools.SortInt64(keys)

		for _, v := range keys {
			rs.Storage.Item = append(rs.Storage.Item, cty.Storage.Content[v])
		}

		keylist := []int{}

		for k := range cty.RessourceProducers {
			keylist = append(keylist, k)
		}

		sort.Ints(keylist)

		for _, k := range keylist {

			var sp simpleProducer
			v := cty.RessourceProducers[k]
			sp.ProducerID = k
			sp.ProductName = v.ProductName
			sp.ProductType = v.ProductType
			sp.Quality = v.Quality
			sp.Quantity = v.Quantity
			_, sp.Active = cty.ActiveRessourceProducers[k]
			if sp.Active {
				sp.EndTime = cty.ActiveRessourceProducers[k].EndTime.Format(time.RFC3339)
			}

			for _, rq := range v.Requirements {
				sp.Requirements += rq.String() + "\n"
			}
			rs.Ressources = append(rs.Ressources, sp)
		}

		keylist = []int{}

		for k := range cty.ProductFactories {
			keylist = append(keylist, k)
		}

		sort.Ints(keylist)

		for _, k := range keylist {
			var sp simpleProducer
			v := cty.ProductFactories[k]
			sp.ProducerID = k
			sp.ProductName = v.ProductName
			sp.ProductType = v.ProductType
			sp.Quality = v.Quality
			sp.Quantity = v.Quantity
			_, sp.Active = cty.ActiveProductFactories[k]
			if sp.Active {
				sp.EndTime = cty.ActiveProductFactories[k].EndTime.Format(time.RFC3339)
			}

			for _, rq := range v.Requirements {
				sp.Requirements += rq.String() + "\n"
			}
			rs.Factories = append(rs.Factories, sp)
		}

		callback <- rs

		if changed {
			dbh := db.New()
			defer dbh.Close()

			cty.Update(dbh)
		}
	})

	res = <-callback

	cb := make(chan simpleNeighbourg)
	defer close(cb)

	for _, v := range res.NeighboursID {
		cm, _ := city_manager.GetCityHandler(v)
		cm.Cast(func(c *city.City) {
			var sn simpleNeighbourg
			sn.ID = c.ID
			sn.Location = c.Location
			sn.Name = c.Name
			cb <- sn
		})
		res.Neighbours = append(res.Neighbours, <-cb)
	}
	return
}

func prepareCity(cities map[int]*city_manager.Handler, cid int) (res simpleCity) {
	callback := make(chan simpleCity)
	defer close(callback)

	neighbours := make(map[int]simpleNeighbourg)
	for v := range cities {
		cb := make(chan simpleNeighbourg)
		defer close(cb)
		cm, _ := city_manager.GetCityHandler(v)
		cm.Cast(func(c *city.City) {
			var sn simpleNeighbourg
			sn.ID = c.ID
			sn.Location = c.Location
			sn.Name = c.Name
			cb <- sn
		})
		neighbours[v] = <-cb
	}

	cities[cid].Cast(func(cty *city.City) {
		res.ID = cty.ID
		res.Name = cty.Name
		res.Location = cty.Location
		for _, v := range cty.NeighboursID {

			res.Neighbours = append(res.Neighbours, neighbours[v])
		}

		callback <- res
	})
	return <-callback
}

func prepareCities(cities map[int]*city_manager.Handler) (res []simpleCity) {
	for k := range cities {
		res = append(res, prepareCity(cities, k))
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

	callback := make(chan map[int]*city_manager.Handler)
	defer close(callback)
	gm.Cast(func(grd *grid.Grid) {
		res := make(map[int]*city_manager.Handler)
		for k := range grd.Cities {
			cm, _ := city_manager.GetCityHandler(k)
			res[k] = cm
		}
		callback <- res
	})

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(prepareCities(<-callback))
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

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(prepareSingleCity(cm))
	} else {
		templates.RenderTemplate(w, req, "city\\show", prepareSingleCity(cm))
	}
}
