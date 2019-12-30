package grid

import (
	"upsilon_cities_go/lib/cities/node"
)

//CompoundedGrid allow to acces to a grid overlay. accessor will provide data based on base + delta grid. Expect delta grid to be initialized with 0 ;)
type CompoundedGrid struct {
	Base  *Grid
	Delta *Grid
}

//IsFilledP tell whether one can work on this location or not.
func (cg CompoundedGrid) IsFilledP(x int, y int) bool {
	return cg.IsFilled(node.NP(x, y))
}

//IsFilled tell whether one can work on this location or not.
func (cg CompoundedGrid) IsFilled(location node.Point) bool {
	n := cg.Base.Get(location).Type
	return n != node.None && n != node.Plain
}

//Get will seek out a node.
func (cg CompoundedGrid) Get(location node.Point) node.Node {
	n := cg.Delta.Get(location)
	if n.Type == node.None {
		return *cg.Base.Get(location)
	}
	return *n
}

//GetP will seek out a node.
func (cg CompoundedGrid) GetP(x int, y int) node.Node {
	n := cg.Delta.GetP(x, y)
	if n.Type == node.None {
		return *cg.Base.GetP(x, y)
	}
	return *n
}

//SetP set value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) SetP(x int, y int, typ node.NodeType) {
	n := cg.Delta.Get(node.NP(x, y))
	if n != nil && n.Type == node.None {
		n.Type = typ
		cg.SetForce(*n)
	}
}

//Set set value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) Set(n node.Node) {
	if cg.Delta.Get(n.Location).Type != node.None {
		cg.SetForce(n)
	}
}

//SetForce set value in delta.
func (cg *CompoundedGrid) SetForce(n node.Node) {
	if n.Type != node.None && !cg.IsFilled(n.Location) {
		nd := cg.Delta.Get(n.Location)
		nd.Type = n.Type
	}
}

//Compact base + delta
func (cg CompoundedGrid) Compact() *Grid {
	for idx, n := range cg.Delta.Nodes {
		if n.Type != node.None {
			cg.Base.Nodes[idx].Type = n.Type
		}
	}
	return cg.Base
}
