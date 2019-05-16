package corporation

import (
	"upsilon_cities_go/lib/misc/generator"
)

type Corporation struct {
	ID       int
	Name     string
	GridID   int
	CitiesID []int

	// user ;)
	OwnerID int
}

//New create a new corporation.
func New(gridID int) (corporation *Corporation) {
	corporation = new(Corporation)
	corporation.GridID = gridID
	corporation.Name = generator.CorpName()
	corporation.OwnerID = 0
	return corporation
}
