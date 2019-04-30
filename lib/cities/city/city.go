package city

import "upsilon_cities_go/lib/cities/node"

type City struct {
	*node.Node // inherit all node

	Neighbours []City
	Roads      []node.Pathway
}
