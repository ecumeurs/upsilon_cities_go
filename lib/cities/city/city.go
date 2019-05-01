package city

import "upsilon_cities_go/lib/cities/node"

type City struct {
	ID         int
	Location   node.Point
	Neighbours []*City
	Roads      []node.Pathway
}
