package caravan_controller

import (
	"encoding/json"
	"log"
	"net/http"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/storage"
	libtools "upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/webtools"
)

//Index GET /caravan List all caravans of current corporation.
func Index(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	cid, err := webtools.CurrentCorpID(req)

	if err != nil {
		webtools.Fail(w, req, "corporation doesn't exist ... maybe it has been kicked from the map", "/map")
		return
	}
	crv, _ := caravan_manager.GetCaravanHandlerByCorpID(cid)

	data := make([]caravan.Caravan, 0)

	for _, v := range crv {
		// Get provide a pointer to storage ... DONT USE IT ;)
		data = append(data, v.Get())
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "caravan/index", data)
	}
}

type candidateCity struct {
	TargetCityID   int
	TargetCityName string
}

type candidateImport struct {
	Item       string
	ItemName   string
	ItemType   []string
	ProducerID int
	ProductID  int
}

type candidateCityExt struct {
	TargetCityID   int
	TargetCityName string
	Imports        []candidateImport
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

	AvailableProducts     []candidate
	JSONAvailableProducts string `json:"-"`

	Cities     []candidateCityExt
	JSONCities string `json:"-"`
}

//New GET /caravan/new/:city_id allow to initiate caravan.
func New(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	cityID, err := webtools.GetInt(req, "city_id")
	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	cm, err := city_manager.GetCityHandler(cityID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
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

			// same we don't care about corporationless cities.
			if city.CorporationID == 0 {
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

			// now prepare cities sellable goods.

			var cce candidateCityExt
			cce.TargetCityID = city.ID
			cce.TargetCityName = city.Name

			for k, v := range city.ProductFactories {
				for kk, vv := range v.Products {
					var ci candidateImport
					ci.Item = vv.StringShort()
					ci.ItemName = vv.ItemName
					ci.ItemType = vv.ItemTypes
					ci.ProducerID = k
					ci.ProductID = kk
					cce.Imports = append(cce.Imports, ci)
				}
			}
			for k, v := range city.RessourceProducers {
				for kk, vv := range v.Products {
					var ci candidateImport
					ci.Item = vv.StringShort()
					ci.ItemName = vv.ItemName
					ci.ItemType = vv.ItemTypes
					ci.ProducerID = k
					ci.ProductID = kk
					cce.Imports = append(cce.Imports, ci)
				}
			}

			data.Cities = append(data.Cities, cce)
		})

	}

	// check sellable ;)

	for k, v := range data.AvailableProducts {
		v.Sellable = len(v.Cities) != 0
		data.AvailableProducts[k] = v
	}

	prods, _ := json.Marshal(data.AvailableProducts)
	cities, _ := json.Marshal(data.Cities)

	data.JSONAvailableProducts = string(prods)
	data.JSONCities = string(cities)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "caravan/new", data)
	}
}

type createJSON struct {
	OriginCityID        int
	TargetCityID        int
	ExportedProducer    int
	ExportedProduct     int
	ExportedMinQuantity int
	ExportedMaxQuantity int
	ExportedMinQuality  int
	ExportedMaxQuality  int
	ImportedProducer    int
	ImportedProduct     int
	ImportedMinQuantity int
	ImportedMaxQuantity int
	ImportedMinQuality  int
	ImportedMaxQuality  int
	OriginExRate        int
	TargetExRate        int
	OriginComp          int
	TargetComp          int
	Delay               int
}

//Create POST /caravan details of caravan. Expect only JS requests on this one ;)
func Create(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}
	if !webtools.CheckAPI(w, req) {
		return
	}

	decoder := json.NewDecoder(req.Body)
	var t createJSON
	err := decoder.Decode(&t)
	if err != nil {
		// Buffer the body
		webtools.Fail(w, req, "unable to parse provided json", "")
		return
	}

	crv := caravan.New()

	crv.CityOriginID = t.OriginCityID
	crv.CityTargetID = t.TargetCityID
	crv.CorpOriginID, _ = webtools.CurrentCorpID(req)

	target, err := city_manager.GetCityHandler(crv.CityTargetID)

	if err != nil {
		webtools.Fail(w, req, "targeted city doesn't exist", "")
		return
	}

	if target.Get().CorporationID == 0 {
		webtools.Fail(w, req, "targeted city doesn't have a corporation", "")
		return
	}

	origin, err := city_manager.GetCityHandler(crv.CityOriginID)

	if err != nil {
		webtools.Fail(w, req, "origin city doesn't exist", "")
		return
	}

	if origin.Get().CorporationID == 0 {
		webtools.Fail(w, req, "origin city doesn't have a corporation", "")
		return
	}

	crv.CorpTargetID = target.Get().CorporationID
	crv.ExchangeRateLHS = t.OriginExRate
	crv.ExchangeRateRHS = t.TargetExRate
	crv.ExportCompensation = t.OriginComp
	crv.ImportCompensation = t.TargetComp

	// seek producer ...
	cb := make(chan *producer.Producer)
	defer close(cb)

	origin.Cast(func(city *city.City) {
		prod, found := city.ProductFactories[t.ExportedProducer]
		if !found {
			prod, found = city.RessourceProducers[t.ExportedProducer]
			if !found {
				cb <- nil
				return
			}
		}

		cb <- prod
	})

	prod := <-cb

	if prod == nil {
		webtools.Fail(w, req, "unable to find requested export producer", "")
		return
	}

	// seek product in producer ;)
	product, found := prod.Products[t.ExportedProduct]
	if !found {
		webtools.Fail(w, req, "unable to find requested exported product", "")
		return
	}

	crv.Exported.ItemType = product.ItemTypes
	crv.Exported.ItemName = product.ItemName
	crv.Exported.Quantity = libtools.IntRange{Min: t.ExportedMinQuantity, Max: t.ExportedMaxQuantity}
	crv.Exported.Quality = libtools.IntRange{Min: t.ExportedMinQuality, Max: t.ExportedMaxQuality}

	target.Cast(func(city *city.City) {
		prod, found := city.ProductFactories[t.ImportedProducer]
		if !found {
			prod, found = city.RessourceProducers[t.ImportedProducer]
			if !found {
				cb <- nil
				return
			}
		}

		cb <- prod
	})

	prod = <-cb

	if prod == nil {
		webtools.Fail(w, req, "unable to find requested imported producer", "")
		return
	}

	// seek product in producer ;)
	product, found = prod.Products[t.ImportedProduct]
	if !found {
		webtools.Fail(w, req, "unable to find requested imported product", "")
		return
	}

	crv.Imported.ItemType = product.ItemTypes
	crv.Imported.ItemName = product.ItemName
	crv.Imported.Quantity = libtools.IntRange{Min: t.ImportedMinQuantity, Max: t.ImportedMaxQuantity}
	crv.Imported.Quality = libtools.IntRange{Min: t.ImportedMinQuality, Max: t.ImportedMaxQuality}
	crv.LoadingDelay = t.Delay
	crv.MapID = origin.Get().MapID

	dbh := db.New()
	defer dbh.Close()

	err = crv.Insert(dbh)

	crv.Reload(dbh) // ensure it get all infos.
	// should load it.

	caravan_manager.GenerateHandler(crv)

	if err != nil {
		log.Printf("CrvCtrl: Failed to insert caravan %+v, %s", crv, err)
		webtools.Fail(w, req, "failed to insert caravan in database", "")
		return
	}

	origin.Call(func(city *city.City) {
		city.Reload(dbh)
	})
	target.Call(func(city *city.City) {
		city.Reload(dbh)
	})

	corp, _ := webtools.CurrentCorp(req)
	corp.Call(func(corp *corporation.Corporation) {
		corp.Reload(dbh)
	})

	targetcorp, _ := corporation_manager.GetCorporationHandler(crv.CorpTargetID)
	targetcorp.Call(func(corp *corporation.Corporation) {
		dbh := db.New()
		defer dbh.Close()
		corp.Reload(dbh)
		if corp.OwnerID == 0 {

			err := crv.Accept(dbh, corp.ID)
			if err != nil {
				log.Printf("CrvCtrl: Failed to auto accept by %d, caravan %+v", corp.ID, crv)
			}
		}
	})

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, "")
	}
}

//Show GET /caravan/:crv_id details of caravan.
func Show(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	crvID, err := webtools.GetInt(req, "crv_id")
	if err != nil {
		webtools.Fail(w, req, "invalid caravan id provided.", "")
		return
	}

	crv, err := caravan_manager.GetCaravanHandler(crvID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	corpID, err := webtools.CurrentCorpID(req)
	_, err = corporation_manager.GetCorporationHandler(corpID)

	if err != nil {
		webtools.Fail(w, req, "can't fetch caravan informations with invalid corporation id", "")
		return
	}

	if crv.Get().CorpOriginID != corpID && crv.Get().CorpTargetID != corpID {
		webtools.Fail(w, req, "can't fetch caravan informations when corporation isn't linked to caravan", "")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(crv.Get())
	} else {
		templates.RenderTemplate(w, req, "caravan/show", crv.Get())
	}
}

//Accept POST /caravan/:crv_id/accept accept new contract
func Accept(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	crvID, err := webtools.GetInt(req, "crv_id")
	if err != nil {
		webtools.Fail(w, req, "invalid caravan id provided.", "")
		return
	}

	crv, err := caravan_manager.GetCaravanHandler(crvID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	corpID, err := webtools.CurrentCorpID(req)
	_, err = corporation_manager.GetCorporationHandler(corpID)

	if err != nil {
		webtools.Fail(w, req, "can't fetch caravan informations with invalid corporation id", "")
		return
	}

	if crv.Get().CorpOriginID != corpID && crv.Get().CorpTargetID != corpID {
		webtools.Fail(w, req, "can't fetch caravan informations when corporation isn't linked to caravan", "")
		return
	}

	cb := make(chan error)

	crv.Cast(func(caravan *caravan.Caravan) {
		dbh := db.New()
		defer dbh.Close()
		cb <- caravan.Accept(dbh, corpID)
	})

	err = <-cb
	if err != nil {
		webtools.Fail(w, req, err.Error(), "")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, "")
	}
}

//Reject POST /caravan/:crv_id/reject reject contract
func Reject(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	crvID, err := webtools.GetInt(req, "crv_id")
	if err != nil {
		webtools.Fail(w, req, "invalid caravan id provided.", "")
		return
	}

	crv, err := caravan_manager.GetCaravanHandler(crvID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	corpID, err := webtools.CurrentCorpID(req)
	_, err = corporation_manager.GetCorporationHandler(corpID)

	if err != nil {
		webtools.Fail(w, req, "can't fetch caravan informations with invalid corporation id", "")
		return
	}

	if crv.Get().CorpOriginID != corpID && crv.Get().CorpTargetID != corpID {
		webtools.Fail(w, req, "can't fetch caravan informations when corporation isn't linked to caravan", "")
		return
	}

	cb := make(chan error)

	crv.Cast(func(caravan *caravan.Caravan) {
		dbh := db.New()
		defer dbh.Close()
		cb <- caravan.Refuse(dbh, corpID)
	})

	err = <-cb
	if err != nil {
		webtools.Fail(w, req, err.Error(), "")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, "")
	}
}

//Abort POST /caravan/:crv_id/abort abort caravan
func Abort(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	crvID, err := webtools.GetInt(req, "crv_id")
	if err != nil {
		webtools.Fail(w, req, "invalid caravan id provided.", "")
		return
	}

	crv, err := caravan_manager.GetCaravanHandler(crvID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	corpID, err := webtools.CurrentCorpID(req)
	_, err = corporation_manager.GetCorporationHandler(corpID)

	if err != nil {
		webtools.Fail(w, req, "can't fetch caravan informations with invalid corporation id", "")
		return
	}

	if crv.Get().CorpOriginID != corpID && crv.Get().CorpTargetID != corpID {
		webtools.Fail(w, req, "can't fetch caravan informations when corporation isn't linked to caravan", "")
		return
	}

	cb := make(chan error)
	defer close(cb)

	crv.Cast(func(caravan *caravan.Caravan) {
		dbh := db.New()
		defer dbh.Close()
		err := caravan.Abort(dbh, corpID)
		log.Printf("CrvCtrl: Aborting: %s %+v", caravan.StringState(corpID), caravan)
		cb <- err
	})

	err = <-cb
	if err != nil {
		webtools.Fail(w, req, err.Error(), "")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, "")
	}
}

type counterData struct {
	ID             int
	OriginCityID   int
	OriginCityName string
	TargetCityID   int
	TargetCityName string

	ExportedItem      string
	ExportedItemTypes string

	ImportedItem      string
	ImportedItemTypes string

	OriginExRate int
	TargetExRate int
	OriginComp   int
	TargetComp   int
	Delay        int
}

//GetCounter GET /caravan/:crv_id/counter propose counter proposition
func GetCounter(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	crvID, err := webtools.GetInt(req, "crv_id")
	if err != nil {
		webtools.Fail(w, req, "invalid caravan id provided.", "")
		return
	}

	crv, err := caravan_manager.GetCaravanHandler(crvID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	corpID, err := webtools.CurrentCorpID(req)
	_, err = corporation_manager.GetCorporationHandler(corpID)

	if err != nil {
		webtools.Fail(w, req, "can't fetch caravan informations with invalid corporation id", "")
		return
	}

	if crv.Get().CorpOriginID != corpID && crv.Get().CorpTargetID != corpID {
		webtools.Fail(w, req, "can't fetch caravan informations when corporation isn't linked to caravan", "")
		return
	}

	if !webtools.CheckLogged(w, req) {
		return
	}

	cb := make(chan counterData)
	defer close(cb)
	crv.Cast(func(caravan *caravan.Caravan) {
		var data counterData

		// do we check exported or imported ... ;)
		data.ID = caravan.ID
		data.OriginCityID = caravan.CityOriginID
		data.TargetCityID = caravan.CityTargetID
		data.OriginCityName = caravan.CityTargetName
		data.TargetCityName = caravan.CityTargetName

		data.ExportedItem = caravan.Exported.String()
		data.ExportedItemTypes = caravan.Exported.StringLong()
		data.ImportedItem = caravan.Imported.String()
		data.ImportedItemTypes = caravan.Imported.StringLong()

		data.OriginExRate = caravan.ExchangeRateLHS
		data.TargetExRate = caravan.ExchangeRateRHS
		data.OriginComp = caravan.ExportCompensation
		data.TargetComp = caravan.ImportCompensation
		data.Delay = caravan.LoadingDelay

		cb <- data

	})

	data := <-cb
	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "caravan/proposition", data)
	}
}

type counterReplyJSON struct {
	ID           int
	OriginExRate int
	TargetExRate int
	OriginComp   int
	TargetComp   int
	Delay        int
}

//PostCounter POST /caravan/:crv_id/counter propose counter proposition
func PostCounter(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	if !webtools.CheckAPI(w, req) {
		return
	}

	crvID, err := webtools.GetInt(req, "crv_id")
	if err != nil {
		webtools.Fail(w, req, "invalid caravan id provided.", "")
		return
	}

	crv, err := caravan_manager.GetCaravanHandler(crvID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	corpID, err := webtools.CurrentCorpID(req)
	_, err = corporation_manager.GetCorporationHandler(corpID)

	if err != nil {
		webtools.Fail(w, req, "can't fetch caravan informations with invalid corporation id", "")
		return
	}

	if crv.Get().CorpOriginID != corpID && crv.Get().CorpTargetID != corpID {
		webtools.Fail(w, req, "can't fetch caravan informations when corporation isn't linked to caravan", "")
		return
	}

	decoder := json.NewDecoder(req.Body)
	var t counterReplyJSON
	err = decoder.Decode(&t)
	if err != nil {
		// Buffer the body
		webtools.Fail(w, req, "unable to parse provided json", "")
		return
	}

	cb := make(chan error)

	crv.Cast(func(caravan *caravan.Caravan) {
		caravan.ExchangeRateLHS = t.OriginExRate
		caravan.ExchangeRateRHS = t.TargetExRate
		caravan.ExportCompensation = t.OriginComp
		caravan.ImportCompensation = t.TargetComp
		caravan.LoadingDelay = t.Delay

		dbh := db.New()
		defer dbh.Close()

		cb <- caravan.Counter(dbh, corpID)
	})

	err = <-cb

	if err != nil {
		webtools.Fail(w, req, err.Error(), "")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, "")
	}
}

//Drop POST /caravan/:crv_id/drop abort and remove from display.
func Drop(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	crvID, err := webtools.GetInt(req, "crv_id")
	if err != nil {
		webtools.Fail(w, req, "invalid caravan id provided.", "")
		return
	}

	crv, err := caravan_manager.GetCaravanHandler(crvID)

	if err != nil {
		webtools.Fail(w, req, "can't initiate caravan creation without an origin city", "")
		return
	}

	corpID, err := webtools.CurrentCorpID(req)
	_, err = corporation_manager.GetCorporationHandler(corpID)

	if err != nil {
		webtools.Fail(w, req, "can't fetch caravan informations with invalid corporation id", "")
		return
	}

	if crv.Get().CorpOriginID != corpID && crv.Get().CorpTargetID != corpID {
		webtools.Fail(w, req, "can't fetch caravan informations when corporation isn't linked to caravan", "")
		return
	}

	cb := make(chan error)
	defer close(cb)

	crv.Cast(func(caravan *caravan.Caravan) {
		dbh := db.New()
		defer dbh.Close()
		err := caravan.CorpDrop(dbh, corpID)
		log.Printf("CrvCtrl: Dropping: %s %+v", caravan.StringState(corpID), caravan)

		// it's already been removed from db by caravan.
		if caravan.OriginDropped && caravan.TargetDropped {
			caravan_manager.DropCaravanHandler(caravan.ID)
		}
		cb <- err
	})

	err = <-cb
	if err != nil {
		webtools.Fail(w, req, err.Error(), "")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, "")
	}
}
