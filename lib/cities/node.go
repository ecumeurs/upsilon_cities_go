package cities

import "math/rand"

type Point struct {
	X int
	Y int
}

type Pathway struct {
	Road       []Point
	FromCityID int
	ToCityID   int
}

type Node struct {
	ID       int
	Location Point
	Type     NodeType
}

//RandomCity assign a random city; the higher scarcity the lower the chance to have a city ;)
func (node *Node) RandomCity(scarcity int) {
	roll := rand.Intn(scarcity + 1)
	if roll < scarcity {
		node.Type = None
	} else {
		node.Type = CityNode
	}
}

func (node *Node) Short() string {
	return node.Type.Short()
}
