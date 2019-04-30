package node

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

func (node *Node) Short() string {
	return node.Type.Short()
}
