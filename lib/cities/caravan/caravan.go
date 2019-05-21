package caravan

import (
	"errors"
	"log"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

//Object describe what will be transited in a Caravan
type Object struct {
	ItemType []string
	Quality  tools.IntRange
	Quantity tools.IntRange
}

// Caravan State
const (
	CRVProposal          int = 0 // Expecting either refused or validated
	CRVCounterProposal   int = 1 // Expecting either refused or validated
	CRVRefused           int = 3
	CRVWaitingOriginLoad int = 4 // waiting at Origin for load
	CRVTravelingToTarget int = 5
	CRVWaitingTargetLoad int = 7 // waiting at Target for load
	CRVTravelingToOrigin int = 8
	CRVAborted           int = 11 // has been aborted
	CRVTerminated        int = 12 // reached end of term
)

//StateToString (state) => string version
var StateToString map[int]string

//StateToDelay (state) => delay (in cycles)
var StateToDelay map[int]int

//StateToNext (state) => (next valid state)
var StateToNext map[int]int

func Init() {
	StateToString = map[int]string{
		CRVProposal:          "Proposal",
		CRVCounterProposal:   "Counter-Proposal",
		CRVRefused:           "Refused",
		CRVWaitingOriginLoad: "Waiting Load",
		CRVTravelingToTarget: "Traveling",
		CRVWaitingTargetLoad: "Waiting Load",
		CRVTravelingToOrigin: "Traveling",
		CRVAborted:           "Aborted",
		CRVTerminated:        "Terminated",
	}
	StateToDelay = map[int]int{
		CRVProposal:          120, // 20 min
		CRVCounterProposal:   120,
		CRVRefused:           0, // Can't move from this state
		CRVWaitingOriginLoad: 0, // this is contract based
		CRVTravelingToTarget: 0,
		CRVWaitingTargetLoad: 0,
		CRVTravelingToOrigin: 0,
		CRVAborted:           0, // Can't move from this state
		CRVTerminated:        0, // Can't move from this state
	}
	StateToNext = map[int]int{
		CRVProposal:          CRVWaitingOriginLoad,
		CRVCounterProposal:   CRVWaitingOriginLoad,
		CRVRefused:           CRVRefused, // Can't move from this state
		CRVWaitingOriginLoad: CRVTravelingToTarget,
		CRVTravelingToTarget: CRVWaitingTargetLoad,
		CRVWaitingTargetLoad: CRVTravelingToOrigin,
		CRVTravelingToOrigin: CRVWaitingOriginLoad,
		CRVAborted:           CRVAborted,    // Can't move from this state
		CRVTerminated:        CRVTerminated, // Can't move from this state
	}
}

//Caravan struct details caravan contract from Contractor POV
type Caravan struct {
	ID    int
	MapID int

	CorpOriginID       int
	CorpOriginName     string
	CityOriginID       int
	CityOriginName     string
	Exported           Object
	ExportCompensation int // money sent with export to buy products.
	SendQty            int

	CorpTargetID       int
	CorpTargetName     string
	CityTargetID       int
	CityTargetName     string
	Imported           Object
	ImportCompensation int // money sent with export to buy products.

	State int

	ExchangeRateLHS int
	ExchangeRateRHS int

	LoadingDelay      int // in cycles
	TravelingDistance int // in nodes
	TravelingSpeed    int // in cycle => this default to 10

	Credits int
	Store   *storage.Storage

	Location node.Point

	Aborted bool

	LastChange time.Time
	NextChange time.Time
	EndOfTerm  time.Time
}

//New instantiate new caravan.
func New() *Caravan {
	cvn := new(Caravan)

	cvn.Store = storage.New()
	cvn.LoadingDelay = 120
	cvn.TravelingDistance = 10
	cvn.TravelingSpeed = 10

	cvn.ExchangeRateLHS = 1
	cvn.ExchangeRateRHS = 1

	cvn.State = CRVProposal
	cvn.LastChange = time.Now().UTC()
	cvn.NextChange = tools.AddCycles(time.Now().UTC(), StateToDelay[CRVWaitingOriginLoad]) // must have refused or counter proposal by 20 min

	return cvn
}

//IsActive Caravan contract is active when not refused, terminated or aborted.
func (caravan *Caravan) IsActive() bool {
	return !(caravan.State == CRVRefused || caravan.State == CRVAborted || caravan.State == CRVTerminated)
}

//IsMoving Caravan contract is active when it's on the road.
func (caravan *Caravan) IsMoving() bool {
	return caravan.IsActive() && (caravan.State == CRVTravelingToOrigin || caravan.State == CRVTravelingToTarget)
}

//IsWaiting Caravan contract is active when it's loading
func (caravan *Caravan) IsWaiting() bool {
	return caravan.IsActive() && (caravan.State == CRVWaitingOriginLoad || caravan.State == CRVWaitingTargetLoad)
}

//Counter caravan contract.
func (caravan *Caravan) Counter(dbh *db.Handler, corporationID int) error {
	if caravan.State == CRVProposal {
		if caravan.State == CRVProposal && caravan.CorpTargetID != corporationID {
			return errors.New("invalid counter")
		}

		caravan.State = CRVCounterProposal
		caravan.LastChange = tools.RoundTime(time.Now().UTC())
		caravan.NextChange = tools.AddCycles(caravan.LastChange, StateToDelay[caravan.State])
		return caravan.Update(dbh)
	}
	return errors.New("invalid state, can't counter")
}

//Refuse caravan contract.
func (caravan *Caravan) Refuse(dbh *db.Handler, corporationID int) error {
	if caravan.State == CRVProposal || caravan.State == CRVCounterProposal {
		if caravan.State == CRVProposal && caravan.CorpTargetID != corporationID {
			return errors.New("invalid refusal")
		}

		if caravan.State == CRVCounterProposal && caravan.CorpOriginID != corporationID {
			return errors.New("invalid refusal")
		}

		caravan.State = CRVRefused
		return caravan.Update(dbh)
	}
	return errors.New("invalid state, can't refuse")
}

//Accept caravan contract.
func (caravan *Caravan) Accept(dbh *db.Handler, corporationID int) error {
	if caravan.State == CRVProposal || caravan.State == CRVCounterProposal {
		if caravan.State == CRVProposal && caravan.CorpTargetID != corporationID {
			return errors.New("invalid accept")
		}

		if caravan.State == CRVCounterProposal && caravan.CorpOriginID != corporationID {
			return errors.New("invalid accept")
		}

		caravan.State = CRVWaitingOriginLoad
		caravan.LastChange = tools.RoundTime(time.Now().UTC())
		caravan.NextChange = tools.AddCycles(caravan.LastChange, StateToDelay[caravan.State])
		return caravan.Update(dbh)
	}
	return errors.New("invalid state, can't accept")
}

//Abort caravan contract. Premature end of contract
func (caravan *Caravan) Abort(dbh *db.Handler) error {
	if caravan.IsActive() {
		caravan.Aborted = true
		return caravan.Update(dbh)
	}
	return errors.New("invalid state, can't refuse")
}

//IsAborted tell whether caravan will soon end or has been ended
func (caravan *Caravan) IsAborted() bool {
	return caravan.Aborted
}

//Fails marks the caravan as a failure due to irrespect of the contract bounds
func (caravan *Caravan) fails() {
	caravan.Aborted = true
	// still need to compensate ...
	caravan.compensate()
}

//compensate ensure that balance of item price is respected.
func (caravan *Caravan) compensate() {
	caravan.Credits = 1000
}

//SetNextState caravan contract.
func (caravan *Caravan) SetNextState(dbh *db.Handler) error {
	if caravan.IsWaiting() {
		if !caravan.IsFilledAtAcceptableLevel() {
			// caravan is going to move but isn't filled.
			caravan.fails()
		} else {
			// caravan is going but may need to compensate in gold.
			caravan.compensate()
		}
		if caravan.State == CRVWaitingOriginLoad {
			items := caravan.Store.All(storage.ByTypesNQuality(caravan.Exported.ItemType, caravan.Exported.Quality))
			count := 0

			for _, v := range items {
				count += v.Quantity
			}

			caravan.SendQty = count
		} else {
			caravan.SendQty = 0
		}
	}

	caravan.State = StateToNext[caravan.State]
	caravan.LastChange = tools.RoundTime(time.Now().UTC())
	if caravan.IsMoving() {
		caravan.NextChange = tools.AddCycles(caravan.LastChange, caravan.TravelingDistance*caravan.TravelingSpeed)
	} else {
		caravan.NextChange = tools.AddCycles(caravan.LastChange, StateToDelay[caravan.State])
	}
	return caravan.Update(dbh)
}

//IsFilled tells whether caravan is ready to go. Works only when caravan is waiting.
func (caravan *Caravan) IsFilled() bool {
	if caravan.State == CRVWaitingOriginLoad {
		// check storage content ...
		items := caravan.Store.All(storage.ByTypesNQuality(caravan.Exported.ItemType, caravan.Exported.Quality))
		count := 0

		for _, v := range items {
			count += v.Quantity
		}

		return count == caravan.Exported.Quantity.Max
	}

	if caravan.State == CRVWaitingTargetLoad {
		// check storage content ...
		items := caravan.Store.All(storage.ByTypesNQuality(caravan.Imported.ItemType, caravan.Exported.Quality))
		count := 0

		for _, v := range items {
			count += v.Quantity
		}

		return count == (caravan.SendQty*caravan.ExchangeRateRHS)/caravan.ExchangeRateLHS
	}
	log.Printf("Caravan: Invalid state")
	return false
}

//IsFilledAtAcceptableLevel tells whether caravan is ready to go. Works only when caravan is waiting.
func (caravan *Caravan) IsFilledAtAcceptableLevel() bool {
	if caravan.State == CRVWaitingOriginLoad {
		// check storage content ...
		items := caravan.Store.All(storage.ByTypesNQuality(caravan.Exported.ItemType, caravan.Exported.Quality))
		count := 0

		for _, v := range items {
			count += v.Quantity
		}

		return count >= caravan.Exported.Quantity.Min
	}

	if caravan.State == CRVWaitingTargetLoad {
		// check storage content ...
		items := caravan.Store.All(storage.ByTypesNQuality(caravan.Imported.ItemType, caravan.Imported.Quality))
		count := 0

		for _, v := range items {
			count += v.Quantity
		}

		return count == (caravan.SendQty*caravan.ExchangeRateRHS)/caravan.ExchangeRateLHS
	}
	log.Printf("Caravan: Invalid state")
	return false
}

//Fill caravan with provided city store.
func (caravan *Caravan) Fill(dbh *db.Handler, city *city.City) error {
	// check first if this is appropriate city to fill from ;)
	var items []item.Item
	max := 0
	if caravan.State == CRVWaitingOriginLoad {
		if caravan.CityOriginID != city.ID {
			return errors.New("Expected to fill from origin city")
		}

		items = city.Storage.All(storage.ByTypesNQuality(caravan.Exported.ItemType, caravan.Exported.Quality))

		max = caravan.Exported.Quantity.Max
	}

	if caravan.State == CRVWaitingTargetLoad {
		if caravan.CityTargetID != city.ID {
			return errors.New("Expected to fill from target city")
		}
		items = city.Storage.All(storage.ByTypesNQuality(caravan.Imported.ItemType, caravan.Imported.Quality))

		max = caravan.Imported.Quantity.Max
	}

	for _, v := range items {
		count := tools.Min(max, v.Quantity)
		max -= count
		city.Storage.Remove(v.ID, count)
		v.Quantity = count
		caravan.Store.Add(v)

		if max == 0 {
			break
		}
	}

	if caravan.State == CRVWaitingOriginLoad {
		caravan.SendQty = caravan.Exported.Quantity.Max - max
	}

	city.Update(dbh)
	caravan.Update(dbh)

	return nil
}

//Unload caravan with provided city store.
func (caravan *Caravan) Unload(dbh *db.Handler, city *city.City) error {
	// check first if this is appropriate city to fill from ;)

	if caravan.State == CRVTravelingToTarget {
		if caravan.CityTargetID != city.ID {
			return errors.New("Expected to unload in target city")
		}
	}

	if caravan.State == CRVTravelingToOrigin {
		if caravan.CityOriginID != city.ID {
			return errors.New("Expected to unload in origin city")
		}
	}

	for _, v := range caravan.Store.Content {

		city.Storage.Add(v)
	}

	caravan.Store.Clear()
	city.Update(dbh)
	caravan.Update(dbh)

	return nil
}
