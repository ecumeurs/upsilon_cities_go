package grid

import (
	"math"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

//CompoundedGrid allow to acces to a grid overlay. accessor will provide data based on base + delta grid. Expect delta grid to be initialized with 0 ;)
type CompoundedGrid struct {
	Base  *Grid
	Delta *Grid
}

//FillSquare will add a square on the grid
func (cg *CompoundedGrid) FillSquare(typ node.NodeType, dist int, centerX int, centerY int) {
	for x := tools.Max(0, centerX-dist); x < tools.Min(cg.Base.Size, centerX+1+dist); x++ {
		for y := tools.Max(0, centerY-dist); y < tools.Min(cg.Base.Size, centerY+1+dist); y++ {
			cg.SetP(x, y, typ)
		}
	}
}

//FillCircle will add a Circle on the grid
func (cg *CompoundedGrid) FillCircle(typ node.NodeType, dist int, centerX int, centerY int) {
	for _, nd := range node.PointsWithinInCircle(node.NP(centerX, centerY), dist, cg.Base.Size) {
		cg.SetP(nd.X, nd.Y, typ)
	}
}

//AddLine will add a Line on the grid
func (cg *CompoundedGrid) AddLine(typ node.NodeType, from node.Point, to node.Point, width int) {

	dist := math.Sqrt(math.Pow(float64(to.X-from.X), 2) + math.Pow(float64(to.Y-from.Y), 2))

	// unit vector = { X/V(X²+Y²), Y/V(X²+Y²) }
	unitX := float64(to.X-from.X) / dist
	unitY := float64(to.Y-from.Y) / dist

	for idx := 0; idx < int(dist); idx++ {
		nd := node.NP(from.X+int(math.Round(unitX*float64(idx))), from.Y+int(math.Round(unitY*float64(idx))))
		cg.SetP(nd.X, nd.Y, typ)
		// now apply Width

		locDist := math.Sqrt(math.Pow(float64(to.X-nd.X), 2) + math.Pow(float64(to.Y-nd.Y), 2))

		locUnitX := -(float64(to.Y-nd.Y) / locDist)
		locUnitY := float64(to.X-nd.X) / locDist
		for widx := -width; widx < width; widx++ {
			nd := node.NP(nd.X+int(math.Round(locUnitX*float64(widx))), nd.Y+int(math.Round(locUnitY*float64(widx))))
			cg.SetP(nd.X, nd.Y, typ)
		}
	}
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
