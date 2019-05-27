package corporation_controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

// No index ;)

type corpInfo struct {
	ID   int
	Name string

	IsOwner  bool
	Extended corpExtended
}

type caravanMeta struct {
	IsDisplayed bool `json:"-"`

	ID             int
	OriginCityID   int
	OriginCityName string

	TargetCityID   int
	TargetCityName string

	IsWaiting         bool
	IsMoving          bool
	IsRequiringAction bool
	IsActive          bool
	CanCounter        bool

	StringState string

	NextUpdate    time.Time
	NextUpdateStr string
}

type corpExtended struct {
	Credits           int
	ActiveCaravans    int
	AvailableCaravans int

	IsViable bool
	Cities   []int

	Caravans []caravanMeta
}

//Show /corporation/:corp_id shows details of corporation
// Will show more if current user is corporation owner.
func Show(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged to access this content.", "")
		return
	}

	reqCorp, _ := tools.GetInt(req, "corp_id")
	corpid, _ := tools.CurrentCorpID(req)

	corp, err := corporation_manager.GetCorporationHandler(reqCorp)
	if err != nil {
		tools.Fail(w, req, "unable to find requested corporation", "")
		return
	}

	cb := make(chan corpInfo)
	defer close(cb)

	corp.Cast(func(corp *corporation.Corporation) {
		var data corpInfo

		data.ID = corp.ID
		data.Name = corp.Name

		if corpid == corp.ID {
			data.IsOwner = true
			data.Extended.Credits = corp.Credits

			storedCrv := make(map[int]bool)

			log.Printf("CorpCtrl: Has %d caravans in stock", len(corp.CaravanID))
			for _, v := range corp.CaravanID {
				if storedCrv[v] {
					// already in.
					continue
				}
				ccb := make(chan caravanMeta)
				storedCrv[v] = true

				defer close(ccb)
				cm, err := caravan_manager.GetCaravanHandler(v)
				if err != nil {
					tools.Fail(w, req, "unable to find caravans information for corporation", "")

					cb <- data
					return
				}

				cm.Cast(func(crv *caravan.Caravan) {
					var meta caravanMeta

					meta.StringState = crv.StringState(corpid)
					meta.IsActive = crv.IsActive()
					meta.IsDisplayed = (!crv.OriginDropped && crv.CorpOriginID == corpid) && (!crv.TargetDropped && crv.CorpTargetID == corpid)
					meta.ID = crv.ID
					meta.OriginCityID = crv.CityOriginID
					meta.OriginCityName = crv.CityOriginName
					meta.TargetCityID = crv.CityTargetID
					meta.TargetCityName = crv.CityTargetName
					meta.IsMoving = crv.IsMoving()
					meta.IsWaiting = crv.IsWaiting()
					meta.IsRequiringAction = (crv.State == caravan.CRVProposal && corpid == crv.CorpTargetID) || (crv.State == caravan.CRVCounterProposal && corpid == crv.CorpOriginID)
					meta.CanCounter = meta.IsRequiringAction && crv.CorpTargetID == corpid
					meta.NextUpdate = crv.NextChange
					meta.NextUpdateStr = crv.NextChange.Format(time.RFC3339)

					ccb <- meta
				})

				mt := <-ccb

				if mt.IsDisplayed {
					data.Extended.Caravans = append(data.Extended.Caravans, mt)
				}

			}

			i := 0

			for _, v := range data.Extended.Caravans {
				if v.IsActive {
					i++
				}
			}

			data.Extended.ActiveCaravans = i
			data.Extended.IsViable = corp.IsViable()
			data.Extended.Cities = corp.CitiesID
		}

		cb <- data
	})

	data := <-cb
	log.Printf("CorpCtrl: About to display corporation: %d as owner? %v", corpid, corpid == reqCorp)
	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(data)
	} else {
		templates.RenderTemplate(w, req, "corporation\\show", data)
	}
}
