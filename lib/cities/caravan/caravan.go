package caravan

import (
	"errors"
	"time"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

//Object describe what will be transited in a Caravan
type Object struct {
	ItemType  []string
	Quality   tools.IntRange
	Quantity  tools.IntRange
	BasePrice int
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

	CorpOriginID   int
	CorpOriginName string
	CityOriginID   int
	CityOriginName string
	Exported       Object

	CorpTargetID   int
	CorpTargetName string
	CityTargetID   int
	CityTargetName string
	Imported       Object

	State int

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

//Refuse caravan contract.
func (caravan *Caravan) Refuse(dbh *db.Handler) error {
	if caravan.State == CRVProposal || caravan.State == CRVCounterProposal {
		caravan.State = CRVRefused
		return caravan.Update(dbh)
	}
	return errors.New("invalid state, can't refuse")
}

//Accept caravan contract.
func (caravan *Caravan) Accept(dbh *db.Handler) error {
	if caravan.State == CRVProposal || caravan.State == CRVCounterProposal {
		caravan.State = CRVWaitingOriginLoad
		caravan.LastChange = tools.RoundTime(time.Now().UTC())
		caravan.NextChange = tools.AddCycles(caravan.LastChange, StateToDelay[caravan.State])
		return caravan.Update(dbh)
	}
	return errors.New("invalid state, can't refuse")
}

//Abort caravan contract. Premature end of contract
func (caravan *Caravan) Abort(dbh *db.Handler) error {
	if caravan.IsActive() {
		caravan.State = CRVWaitingOriginLoad
		caravan.LastChange = tools.RoundTime(time.Now().UTC())
		caravan.NextChange = tools.AddCycles(caravan.LastChange, StateToDelay[caravan.State])
		return caravan.Update(dbh)
	}
	return errors.New("invalid state, can't refuse")
}

//SetNextState caravan contract.
func (caravan *Caravan) SetNextState(dbh *db.Handler) error {
	caravan.State = StateToNext[caravan.State]
	caravan.LastChange = tools.RoundTime(time.Now().UTC())
	if caravan.State == CRVTravelingToOrigin || caravan.State == CRVTravelingToTarget {
		caravan.NextChange = tools.AddCycles(caravan.LastChange, caravan.TravelingDistance*caravan.TravelingSpeed)
	} else {
		caravan.NextChange = tools.AddCycles(caravan.LastChange, StateToDelay[caravan.State])
	}
	return caravan.Update(dbh)
}
