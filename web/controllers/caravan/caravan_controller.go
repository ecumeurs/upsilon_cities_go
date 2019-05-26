package caravan_controller

import (
	"encoding/json"
	"net/http"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/storage"
	libtools "upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

//Index GET /caravan List all caravans of current corporation.
func Index(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckLogged(w, req) {
		return
	}

	cid, err := tools.CurrentCorpID(req)

	if err != nil {
		tools.Fail(w, req, "corporation doesn't exist ... maybe it has been kicked from the map", "/map")
		return
	}
	crv, _ := caravan_manager.GetCaravanHandlerByCorpID(cid)

	data := make([]caravan.Caravan, 0)

	for _, v := range crv {
		// Get provide a pointer to storage ... DONT USE IT ;)
		data = append(data, v.Get())
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "caravan\\index", data)
	}
}

type candidateCity struct {
	TargetCityID   int
	TargetCityName string
}

type candidate struct {
	Item       string
	ItemName   string
	ItemType   []string
	ProducerID int
	ProductID  int

	Production         int
	ProductionQuality  libtools.IntRange
	ProductionDuration int

	Cities              []candidateCity
	Sellable            bool // if no cities, then ain't sellable ... ;)
	AlreadyExchanged    bool // if there is already a caravan using it.
	HasRecentlyProduced bool // if hasn't produced last production duration x 2 then ... might be chancy to create a caravan on it ;)
	CurrentStock        int
}

type newData struct {
	OriginCityID   int
	OriginCityName string

	AvailableProducts []candidate
}

//New GET /caravan/new/:city_id allow to initiate caravan.
func New(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckLogged(w, req) {
		return
	}

	cityID, err := tools.GetInt(req, "city_id")
	if err != nil {
		tools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	cm, err := city_manager.GetCityHandler(cityID)

	if err != nil {
		tools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	cb := make(chan newData)
	defer close(cb)

	neighbours := cm.Get().NeighboursID

	cm.Cast(func(city *city.City) {
		var data newData
		data.OriginCityID = city.ID
		data.OriginCityName = city.Name

		for k, v := range city.RessourceProducers {
			for kk, vv := range v.Products {
				var cd candidate
				cd.ProducerID = k
				cd.ProductID = kk
				cd.Production = vv.Quantity.Min
				cd.ProductionQuality = vv.Quality
				cd.ProductionDuration = v.Delay
				cd.CurrentStock = city.Storage.CountAll(storage.ByTypesNQuality(vv.ItemTypes, vv.Quality))
				cd.HasRecentlyProduced = libtools.AboutNow(v.Delay * -2).Before(v.LastActivity)
				cd.Item = vv.StringShort()
				cd.ItemType = vv.ItemTypes
				cd.ItemName = vv.ItemName

				// fills candidate cities & already exchanged later ( avoids locks within locks)
				data.AvailableProducts = append(data.AvailableProducts, cd)
			}
		}

		for k, v := range city.ProductFactories {
			for kk, vv := range v.Products {
				var cd candidate
				cd.ProducerID = k
				cd.ProductID = kk
				cd.Production = vv.Quantity.Min
				cd.ProductionQuality = vv.Quality
				cd.ProductionDuration = v.Delay
				cd.CurrentStock = city.Storage.CountAll(storage.ByTypesNQuality(vv.ItemTypes, vv.Quality))
				cd.HasRecentlyProduced = libtools.AboutNow(v.Delay * -2).Before(v.LastActivity)
				cd.Item = vv.StringShort()
				cd.ItemType = vv.ItemTypes
				cd.ItemName = vv.ItemName

				// fills candidate cities & already exchanged later ( avoids locks within locks)
				data.AvailableProducts = append(data.AvailableProducts, cd)
			}
		}

		cb <- data
	})

	data := <-cb

	cms, _ := city_manager.GetCityHandlerByMap(cm.Get().MapID)
	crvs, _ := caravan_manager.GetCaravanHandlerByCityID(cityID)

	for _, crv := range crvs {
		// call ensure we wait for end of this function before continuing on.
		// as we alter "data" that live in this thread (and not in caravan thread)
		crv.Call(func(caravan *caravan.Caravan) {

			// do we check exported or imported ... ;)
			exported := caravan.CityOriginID == cityID

			for k, cd := range data.AvailableProducts {
				if exported {
					if libtools.ListInStringList(caravan.Exported.ItemType, cd.ItemType) {
						cd.AlreadyExchanged = true
						data.AvailableProducts[k] = cd
						break
					}
				} else {
					if libtools.ListInStringList(caravan.Imported.ItemType, cd.ItemType) {
						cd.AlreadyExchanged = true
						data.AvailableProducts[k] = cd
						break
					}
				}
			}
		})
	}

	for _, cty := range cms {
		// call ensure we wait for end of this function before continuing on.
		// as we alter "data" that live in this thread (and not in city thread)

		if !libtools.InList(cty.Get().ID, neighbours) {
			continue
		}

		cty.Call(func(city *city.City) {
			// of course we don't care about our city ;)
			if city.ID == cityID {
				return
			}

			for k, cd := range data.AvailableProducts {
				// target city won't by anything that it produces already

				hasProducer := false

				for _, v := range city.RessourceProducers {
					for _, vv := range v.Products {
						if libtools.ListInStringList(cd.ItemType, vv.ItemTypes) {
							hasProducer = true
							break
						}
					}
					if hasProducer {
						break
					}
				}
				if !hasProducer {
					for _, v := range city.ProductFactories {
						for _, vv := range v.Products {
							if libtools.ListInStringList(cd.ItemType, vv.ItemTypes) {
								hasProducer = true
								break
							}
						}
						if hasProducer {
							break
						}
					}
				}

				if !hasProducer {
					var ccity candidateCity
					ccity.TargetCityID = city.ID
					ccity.TargetCityName = city.Name

					cd.Cities = append(cd.Cities, ccity)

					data.AvailableProducts[k] = cd
				}
			}
		})
	}

	// check sellable ;)

	for k, v := range data.AvailableProducts {
		v.Sellable = len(v.Cities) != 0
		data.AvailableProducts[k] = v
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "caravan\\new", data)
	}
}

//Create POST /caravan details of caravan. Expect only JS requests on this one ;)
func Create(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckLogged(w, req) {
		return
	}
	if !tools.CheckAPI(w, req) {
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.Redirect(w, req, "")
	}
}

//Seek GET /caravan/seek seek cities candidate for provided items.
func Seek(w http.ResponseWriter, req *http.Request) {

}

//Show GET /caravan/:crv_id details of caravan.
func Show(w http.ResponseWriter, req *http.Request) {

}

//Accept POST /caravan/:crv_id/accept accept new contract
func Accept(w http.ResponseWriter, req *http.Request) {

}

//Reject POST /caravan/:crv_id/reject reject contract
func Reject(w http.ResponseWriter, req *http.Request) {

}

//Abort POST /caravan/:crv_id/abort abort caravan
func Abort(w http.ResponseWriter, req *http.Request) {

}

//GetCounter GET /caravan/:crv_id/counter propose counter proposition
func GetCounter(w http.ResponseWriter, req *http.Request) {

}

//PostCounter POST /caravan/:crv_id/counter propose counter proposition
func PostCounter(w http.ResponseWriter, req *http.Request) {

}

//Drop POST /caravan/:crv_id/drop abort and remove from display.
func Drop(w http.ResponseWriter, req *http.Request) {

}
