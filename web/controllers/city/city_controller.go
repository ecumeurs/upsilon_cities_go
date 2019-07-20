//Package city_controller will allow manipulation of Cities and information gathering ...
package city_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/grid"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/node"
	lib_tools "upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/gameplay"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

type simpleNeighbourg struct {
	ID       int
	Location node.Point
	Name     string
}

type simpleItem struct {
	Items []item.Item
	Count int
	Name  string
	Types string
	IDStr string `json:"-"` // helper for collapser
}

type simpleStorage struct {
	Count    int
	Capacity int
	Item     map[string]simpleItem // items by name
}

type simpleProduct struct {
	ID          int
	ProductType []string
	ProductName string
	Quality     lib_tools.IntRange
	Quantity    lib_tools.IntRange
	Upgrade     bool
	BigUpgrade  bool
}

type simpleProducer struct {
	ProducerID   int
	ProducerName string
	Products     []simpleProduct
	Active       bool
	EndTime      string
	Requirements string
	BigUpgrade   bool
	CityID       int
	Owner        bool
}

type simpleCaravan struct {
	ID               int
	To               bool   // tell whether this caravan originate from this city or not.
	ExportedItem     string // item designation.
	ImportedItem     string // item designation.
	ExportedItemLong string // item designation.
	ImportedItemLong string // item designation.
	CityName         string // remote city name

	Active    bool
	Displayed bool `json:"-"`
}

type simpleCity struct {
	ID              int
	Location        node.Point
	Neighbours      []simpleNeighbourg
	NeighboursID    []int
	Name            string
	CorporationName string
	CorpoID         int
	Filled          bool
	Fame            int
	Storage         simpleStorage
	Ressources      []simpleProducer
	Factories       []simpleProducer
	Caravans        []simpleCaravan
}

type upgrade struct {
	CityID  int
	Message int
	Result  bool
}

func prepareSingleCity(corpID int, cm *city_manager.Handler) (res simpleCity) {
	callback := make(chan simpleCity)
	defer close(callback)

	cm.Cast(func(cty *city.City) {
		changed := cty.CheckActivity(time.Now().UTC())
		if changed {
			dbh := db.New()
			defer dbh.Close()
			cty.CheckCityOwnership(dbh)
			// might have removed corporation from it ;)
		}
		var rs simpleCity
		rs.ID = cty.ID
		rs.Name = cty.Name
		rs.Location = cty.Location
		rs.NeighboursID = cty.NeighboursID
		rs.CorpoID = cty.CorporationID
		rs.CorporationName = cty.CorporationName
		rs.Filled = cty.CorporationID == corpID
		rs.Fame = cty.Fame[corpID]

		log.Printf("City: Preping city for display targeted corp %d city fame %v found fame %d", corpID, cty.Fame, rs.Fame)

		keylist := []int{}

		for k := range cty.RessourceProducers {
			keylist = append(keylist, k)
		}

		sort.Ints(keylist)

		for _, k := range keylist {

			var sp simpleProducer
			v := cty.RessourceProducers[k]
			sp.ProducerID = k
			sp.CityID = rs.ID
			sp.Owner = cty.CorporationID == corpID
			sp.ProducerName = v.Name
			for _, w := range v.Products {
				var p simpleProduct
				p.ID = w.ID
				p.ProductName = w.ItemName
				p.ProductType = w.ItemTypes
				p.Quality = w.GetQuality()
				p.Quantity = w.GetQuantity()
				p.BigUpgrade = v.CanBigUpgrade()
				p.Upgrade = v.CanUpgrade()
				sp.Products = append(sp.Products, p)
			}
			sp.BigUpgrade = v.CanBigUpgrade()

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
			sp.ProducerName = v.Name
			sp.CityID = rs.ID
			sp.Owner = cty.CorporationID == corpID
			for _, w := range v.Products {
				var p simpleProduct
				p.ID = w.ID
				p.ProductName = w.ItemName
				p.ProductType = w.ItemTypes
				p.Quality = w.GetQuality()
				p.Quantity = w.GetQuantity()
				p.BigUpgrade = v.CanBigUpgrade()
				p.Upgrade = v.CanUpgrade()
				sp.Products = append(sp.Products, p)
			}
			sp.BigUpgrade = v.CanBigUpgrade()

			_, sp.Active = cty.ActiveProductFactories[k]
			if sp.Active {
				sp.EndTime = cty.ActiveProductFactories[k].EndTime.Format(time.RFC3339)
			}

			for _, rq := range v.Requirements {
				sp.Requirements += rq.String() + "\n"
			}
			rs.Factories = append(rs.Factories, sp)
		}

		if cty.CorporationID == corpID {

			rs.Storage.Count = cty.Storage.Count()
			rs.Storage.Capacity = cty.Storage.Capacity
			rs.Storage.Item = make(map[string]simpleItem)

			// To store the keys in slice in sorted order
			var keys []int64
			for k := range cty.Storage.Content {
				keys = append(keys, k)
			}

			lib_tools.SortInt64(keys)

			for _, v := range keys {
				itm := cty.Storage.Content[v]
				var sitm simpleItem
				sitm, found := rs.Storage.Item[itm.Name]
				if !found {
					sitm.Name = itm.Name
					sitm.IDStr = strings.ReplaceAll(strings.ReplaceAll(sitm.Name, " ", "-"), "'", "-")
					sitm.Types = itm.PrettyTypes()
				}

				sitm.Count += itm.Quantity
				sitm.Items = append(sitm.Items, itm)

				rs.Storage.Item[itm.Name] = sitm
			}

			knownCaravans := make(map[int]bool)

			for _, v := range cty.CaravanID {
				ccb := make(chan simpleCaravan)
				defer close(ccb)

				crvm, _ := caravan_manager.GetCaravanHandler(v)
				crvm.Cast(func(caravan *caravan.Caravan) {
					var res simpleCaravan

					res.ID = caravan.ID
					res.To = caravan.CityOriginID == rs.ID
					if res.To {
						res.ExportedItem = caravan.Exported.String()
						res.ImportedItem = caravan.Imported.String()
						res.CityName = caravan.CityTargetName
						res.Displayed = !caravan.OriginDropped
					} else {
						res.ExportedItem = caravan.Imported.String()
						res.ImportedItem = caravan.Exported.String()
						res.CityName = caravan.CityOriginName
						res.Displayed = !caravan.TargetDropped
					}

					res.Active = caravan.IsActive()

					ccb <- res
				})

				dt := <-ccb

				if !knownCaravans[dt.ID] {
					if dt.Displayed {
						rs.Caravans = append(rs.Caravans, dt)
					}
					knownCaravans[dt.ID] = true
				}
			}
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

// Show GET: /city/:city_id
func Show(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckLogged(w, req) {
		return
	}

	cityID, err := tools.GetInt(req, "city_id")

	cm, err := city_manager.GetCityHandler(cityID)
	if err != nil {
		tools.Fail(w, req, "Unknown city id", "")
		return
	}

	gm, _ := grid_manager.GetGridHandler(cm.Get().MapID)

	// call on actor, so we don't get into a running region update.
	updateNeeded := make(chan bool)
	defer close(updateNeeded)
	gm.Cast(func(grid *grid.Grid) {
		updateNeeded <- grid.RegionUpdateNeeded(cityID)
	})

	// this should ensure caravan evolution.
	if <-updateNeeded {
		// we need to wait for this to finish before proceeding.
		gm.Call(func(grid *grid.Grid) {
			grid.UpdateRegion()
		})
	}

	corpid, _ := tools.CurrentCorpID(req)

	log.Printf("CityCtrl: About to display city: %d as corp %d", cityID, corpid)
	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(prepareSingleCity(corpid, cm))
	} else {
		templates.RenderTemplate(w, req, "city/show", prepareSingleCity(corpid, cm))
	}
}

//ProducerUpgrade update Producer depending on user chose
func ProducerUpgrade(w http.ResponseWriter, req *http.Request) {
	cityID, err := tools.GetInt(req, "city_id")
	producerID, err := tools.GetInt(req, "producer_id")
	action, err := tools.GetInt(req, "action")
	product, err := tools.GetInt(req, "product")

	cm, err := city_manager.GetCityHandler(cityID)
	if err != nil {
		tools.Fail(w, req, "Unknown city id", "")
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(upgradeSingleProducer(cm, producerID, action, product))
	}
}

//ProducerUpgrade update Producer depending on user chose
func upgradeSingleProducer(cm *city_manager.Handler, prodID int, actionID int, product int) (res upgrade) {

	callback := make(chan upgrade)
	defer close(callback)

	cm.Cast(func(cty *city.City) {
		var changed bool

		if val, ok := cty.ProductFactories[prodID]; ok {
			changed = val.Upgrade(actionID, product)
		}

		if val, ok := cty.RessourceProducers[prodID]; ok {
			changed = val.Upgrade(actionID, product)
		}

		var rs upgrade
		rs.CityID = cty.ID
		rs.Result = changed

		callback <- rs

		if changed {
			dbh := db.New()
			defer dbh.Close()
			cty.Update(dbh)
		}
	})

	res = <-callback

	return
}

type itemOpRes struct {
	Item       item.Item
	Producable bool
	Success    bool
}

//Give POST /city/:city_id/give/:item
func Give(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckLogged(w, req) {
		return
	}

	corpid, err := tools.CurrentCorpID(req)
	if err != nil {
		tools.Fail(w, req, "unable to find corporation ... can't proceed", "/map")
		return
	}

	cityID, err := tools.GetInt(req, "city_id")

	cm, err := city_manager.GetCityHandler(cityID)
	if err != nil {
		tools.Fail(w, req, "Unknown city id", "")
		return
	}

	itm, err := tools.GetInt(req, "item")
	if err != nil {
		tools.Fail(w, req, "unable to parse requested item", "")
		return
	}

	if !cm.Get().Storage.Has(int64(itm)) {
		tools.Fail(w, req, "requested item isn't in store", "")
		return
	}

	cb := make(chan itemOpRes)
	defer close(cb)
	// do your magic here

	cm.Cast(func(city *city.City) {
		var r itemOpRes
		if !city.Storage.Has(int64(itm)) {
			r.Success = false
			cb <- r
			return
		}
		r.Item = city.Storage.Content[int64(itm)]
		r.Success = true
		r.Producable = city.CanProduce(r.Item)
		city.Storage.Remove(int64(itm), 0) // remove all

		if r.Producable {
			city.AddFame(corpid, fmt.Sprintf("gave %s to the city", r.Item.Name), lib_tools.Floor(float32(r.Item.Price()*r.Item.Quantity)*gameplay.GetFloat("producable_item_fame", 0.1)))
		} else {
			city.AddFame(corpid, fmt.Sprintf("gave %s to the city", r.Item.Name), lib_tools.Floor(float32(r.Item.Price()*r.Item.Quantity)*gameplay.GetFloat("unproducable_item_fame", 0.1)))
		}

		cb <- r
	})

	opres := <-cb

	if opres.Success {

		log.Printf("CityCtrl: About to display city: %d as corp %d", cityID, corpid)
		if tools.IsAPI(req) {
			tools.GenerateAPIOk(w)
			json.NewEncoder(w).Encode(opres.Item)
		} else {
			templates.RenderTemplate(w, req, "city/item", opres.Item)
		}
	} else {
		tools.Fail(w, req, "fail to perform operation", "/map")
	}
}

//Drop POST /city/:city_id/sell/:item
func Drop(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckLogged(w, req) {
		return
	}

	corpid, err := tools.CurrentCorpID(req)
	if err != nil {
		tools.Fail(w, req, "unable to find corporation ... can't proceed", "/map")
		return
	}

	cityID, err := tools.GetInt(req, "city_id")

	cm, err := city_manager.GetCityHandler(cityID)
	if err != nil {
		tools.Fail(w, req, "Unknown city id", "")
		return
	}

	itm, err := tools.GetInt(req, "item")
	if err != nil {
		tools.Fail(w, req, "unable to parse requested item", "")
		return
	}

	if !cm.Get().Storage.Has(int64(itm)) {
		tools.Fail(w, req, "requested item isn't in store", "")
		return
	}

	cb := make(chan itemOpRes)
	defer close(cb)
	// do your magic here

	cm.Cast(func(city *city.City) {
		var r itemOpRes
		if !city.Storage.Has(int64(itm)) {
			r.Success = false
			cb <- r
			return
		}
		r.Item = city.Storage.Content[int64(itm)]
		r.Success = true
		city.Storage.Remove(int64(itm), 0) // remove all
		cb <- r
	})

	opres := <-cb

	if opres.Success {

		log.Printf("CityCtrl: About to display city: %d as corp %d", cityID, corpid)
		if tools.IsAPI(req) {
			tools.GenerateAPIOk(w)
			json.NewEncoder(w).Encode(opres.Item)
		} else {
			templates.RenderTemplate(w, req, "city/item", opres.Item)
		}
	} else {
		tools.Fail(w, req, "fail to perform operation", "/map")
	}
}

//Sell POST /city/:city_id/sell/:item
func Sell(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckLogged(w, req) {
		return
	}

	corpid, err := tools.CurrentCorpID(req)
	if err != nil {
		tools.Fail(w, req, "unable to find corporation ... can't proceed", "/map")
		return
	}

	cityID, err := tools.GetInt(req, "city_id")

	cm, err := city_manager.GetCityHandler(cityID)
	if err != nil {
		tools.Fail(w, req, "Unknown city id", "")
		return
	}

	itm, err := tools.GetInt(req, "item")
	if err != nil {
		tools.Fail(w, req, "unable to parse requested item", "")
		return
	}

	if !cm.Get().Storage.Has(int64(itm)) {
		tools.Fail(w, req, "requested item isn't in store", "")
		return
	}

	cb := make(chan itemOpRes)
	defer close(cb)
	// do your magic here

	cm.Cast(func(city *city.City) {
		var r itemOpRes
		if !city.Storage.Has(int64(itm)) {
			r.Success = false
			cb <- r
			return
		}
		r.Item = city.Storage.Content[int64(itm)]
		r.Success = true
		r.Producable = city.CanProduce(r.Item)
		city.Storage.Remove(int64(itm), 0) // remove all
		cb <- r
	})

	opres := <-cb

	if opres.Success {

		corpm, _ := tools.CurrentCorp(req)

		corpm.Call(func(corp *corporation.Corporation) {
			if opres.Producable {
				corp.Credits += lib_tools.Floor(float32(opres.Item.Price()*opres.Item.Quantity) * gameplay.GetFloat("producable_item_price", 0.5))
			} else {
				corp.Credits += lib_tools.Floor(float32(opres.Item.Price()*opres.Item.Quantity) * gameplay.GetFloat("unproducable_item_price", 1))
			}
		})

		log.Printf("CityCtrl: About to display city: %d as corp %d", cityID, corpid)
		if tools.IsAPI(req) {
			tools.GenerateAPIOk(w)
			json.NewEncoder(w).Encode(opres.Item)
		} else {
			templates.RenderTemplate(w, req, "city/item", opres.Item)
		}
	} else {
		tools.Fail(w, req, "fail to perform operation", "/map")
	}
}
