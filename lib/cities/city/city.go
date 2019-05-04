package city

import (
	"upsilon_cities_go/lib/cities/node"
)

//City
type City struct {
	ID         int
	Name       string
	Location   node.Point
	Neighbours []*City
	Roads      []node.Pathway
}