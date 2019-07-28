package city_evolution

import (
	"errors"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/misc/config/gameplay"
)

// City State
const (
	CSCaravan        int = 0
	CSRessources     int = 1
	CSFactories      int = 2
	CSResellers      int = 3
	CSStorage        int = 4
	CSProductionRate int = 5
)

var stateToString map[int]string

var stateToUpgrade map[int]func(*city.State)

//CSInit initialiaze cities states.
func CSInit() {
	stateToString = map[int]string{
		CSCaravan:        "Caravan",
		CSRessources:     "Ressources",
		CSFactories:      "Factories",
		CSResellers:      "Resellers",
		CSStorage:        "Storage",
		CSProductionRate: "ProductionRate",
	}

	stateToUpgrade = map[int]func(*city.State){
		CSCaravan: func(state *city.State) {
			state.MaxCaravans++
		},

		CSRessources: func(state *city.State) {
			state.MaxRessources++
		},

		CSFactories: func(state *city.State) {
			state.MaxFactories++
		},

		CSResellers: func(state *city.State) {
			state.MaxResellers++
		},

		CSStorage: func(state *city.State) {
			state.MaxStorageSpace = 100 + state.MaxStorageSpace
		},

		CSProductionRate: func(state *city.State) {
			state.ProductionRate = state.ProductionRate + 0.1
		},
	}
}

//Init a state
func Init(state *city.State) {
	state.CurrentLevel = 1
	state.MaxCaravans = gameplay.GetInt("init_city_max_caravan", 3)
	state.MaxRessources = gameplay.GetInt("init_city_max_ressources", 3)
	state.MaxFactories = gameplay.GetInt("init_city_max_factories", 3)
	state.MaxResellers = gameplay.GetInt("init_city_max_resellers", 3)
	state.MaxStorageSpace = gameplay.GetInt("init_city_storage_space", 3)
	state.ProductionRate = gameplay.GetFloat("init_city_production_rate", 3)
}

//NextLevelRequirements specify what's required to perform a level up.
func NextLevelRequirements(state city.State) (credits, fame, ressources int) {
	// atm upgrade is strictly identical, no matter what, but should aim at having an exponentialish kind of curve.
	// also ressources requirements should be actual items ;) and not simply a number of ressource.
	return 1000, 500, 200
}

//LevelUp upgrades current city state.
func LevelUp(state *city.State, update int) error {
	if _, found := stateToString[update]; !found {
		return errors.New("unable to perform levelup as request upgrade isn't available")
	}

	var sh city.StateHistory
	sh.Date = time.Now().UTC()
	sh.IncreaseType = update
	sh.Level = state.CurrentLevel + 1
	state.CurrentLevel++
	stateToUpgrade[update](state)
	state.History = append(state.History, sh)
	return nil
}
