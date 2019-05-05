package city

import (
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/storage"
)

//City
type City struct {
	ID            int
	Location      node.Point
	NeighboursID  []int
	Roads         []node.Pathway
	Name          string
	CorporationID int
	Storage       storage.Storage
}
