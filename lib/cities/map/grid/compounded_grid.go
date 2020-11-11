package grid

import (
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
)

//CompoundedGrid allow to acces to a grid overlay. accessor will provide data based on base + delta grid. Expect delta grid to be initialized with 0 ;)
type CompoundedGrid struct {
	Base  *Grid
	Delta *Grid
}

//IsFilledP tell whether one can work on this location or not.
func (cg CompoundedGrid) IsFilledP(x int, y int, change nodetype.ChangeType) bool {
	return cg.IsFilled(node.NP(x, y), change)
}

//IsFilled tell whether one can work on this location or not.
func (cg CompoundedGrid) IsFilled(location node.Point, change nodetype.ChangeType) bool {
	n := cg.Base.Get(location)
	return cg.NIsFilled(*n, change)
}

//IsDeltaFilled tell whether one can work on this location or not (in this delta).
func (cg CompoundedGrid) IsDeltaFilled(location node.Point, change nodetype.ChangeType) bool {
	n := cg.Delta.Get(location)
	return cg.NIsFilled(*n, change)
}

//NIsFilled tell whether one can work on this location or not.
func (cg CompoundedGrid) NIsFilled(n node.Node, change nodetype.ChangeType) bool {
	switch change {
	case nodetype.Any:
		return n.Type != nodetype.None || n.Ground != cg.Base.Base || n.Landscape != nodetype.NoLandscape || n.IsStructure || n.IsRoad
	case nodetype.Type:
		return n.Type != nodetype.None
	case nodetype.Ground:
		return n.Ground != cg.Base.Base
	case nodetype.Landscape:
		return n.Landscape != nodetype.NoLandscape
	case nodetype.Structure:
		return n.IsStructure
	case nodetype.Road:
		return n.IsRoad
	default:
		return n.Type == nodetype.Filled
	}
}

//Get will seek out a node.
func (cg CompoundedGrid) Get(location node.Point) node.Node {
	n := cg.Delta.Get(location)
	if n.Type == nodetype.None {
		return *cg.Base.Get(location)
	}
	return *n
}

//GetP will seek out a node.
func (cg CompoundedGrid) GetP(x int, y int) node.Node {
	n := *cg.Delta.GetP(x, y)
	if !cg.NIsFilled(n, nodetype.Any) {
		return *cg.Base.GetP(x, y)
	}
	return n
}

//SetPNT set NodeType value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) SetPNT(x int, y int, typ nodetype.NodeType) {
	n := cg.Delta.Get(node.NP(x, y))
	if n != nil && n.Type == nodetype.None {
		nn := *n
		nn.Type = typ
		cg.Set(nn, nodetype.Type)
	}
}

//SetPGT set NodeType value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) SetPGT(x int, y int, typ nodetype.GroundType) {
	n := cg.Delta.Get(node.NP(x, y))
	if n != nil && n.Type == nodetype.None {
		nn := *n
		nn.Ground = typ
		cg.Set(nn, nodetype.Ground)
	}
}

//SetPLT set LandscapeType value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) SetPLT(x int, y int, typ nodetype.LandscapeType) {
	n := cg.Delta.Get(node.NP(x, y))
	if n != nil && n.Type == nodetype.None {
		nn := *n
		nn.Landscape = typ
		cg.Set(nn, nodetype.Landscape)
	}
}

//SetPRoad set Road value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) SetPRoad(x int, y int, road bool) {
	n := cg.Delta.Get(node.NP(x, y))
	if n != nil && n.Type == nodetype.None {
		nn := *n
		nn.IsRoad = road
		cg.Set(nn, nodetype.Road)
	}
}

//SetPCity set LandscapeType value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) SetPCity(x int, y int, cty bool) {
	n := cg.Delta.Get(node.NP(x, y))
	if n != nil && n.Type == nodetype.None {
		nn := *n
		nn.IsStructure = cty
		cg.Set(nn, nodetype.Structure)
	}
}

//Set set value in delta, if there is nothing in delta.
func (cg *CompoundedGrid) Set(n node.Node, change nodetype.ChangeType) {
	if !cg.IsDeltaFilled(n.Location, change) {
		nd := cg.Delta.Get(n.Location)
		nd.Update(&n)
		nd.Type = nodetype.Filled
	}
}

//SetForce set value in delta.
func (cg *CompoundedGrid) SetForce(n node.Node) {
	nd := cg.Delta.Get(n.Location)
	nd.Update(&n)
	nd.Type = nodetype.Filled
}

//Update set value in delta.
func (cg *CompoundedGrid) Update(n node.Node) {
	nd := cg.Delta.Get(n.Location)
	nd.Update(&n)
	nd.Type = nodetype.Filled
}

//Compact base + delta
func (cg *CompoundedGrid) Compact() *Grid {
	for idx, n := range cg.Delta.Nodes {
		if n.Type == nodetype.Filled {
			cg.Base.Nodes[idx].Update(&n)
		}
	}

	for k, v := range cg.Delta.Cities {
		cg.Base.Cities[k] = v
	}

	for k, v := range cg.Delta.LocationToCity {
		cg.Base.LocationToCity[k] = v
	}

	cg.Delta = Create(cg.Base.Size, cg.Base.Base)

	return cg.Base
}

//SelectPattern will select corresponding nodes in a grid based on pattern & location
func (cg *CompoundedGrid) SelectPattern(loc node.Point, pattern pattern.Pattern) []node.Node {
	res := make([]node.Node, 0, len(pattern))
	for _, v := range pattern.Apply(loc, cg.Base.Size) {
		res = append(res, cg.Get(v))
	}
	return res
}

//SelectPatternMapBorders will select corresponding nodes in a grid based on pattern & location
func (cg *CompoundedGrid) SelectPatternMapBorders(loc node.Point, pattern pattern.Pattern) []node.Node {
	res := make([]node.Node, 0, len(pattern))
	for _, v := range pattern.ApplyBorders(loc, cg.Base.Size) {
		res = append(res, cg.Get(v))
	}
	return res
}

//SelectMapBorders will retrieve nodes for map borders.
func (cg *CompoundedGrid) SelectMapBorders() []node.Node {
	res := make([]node.Node, 0, cg.Base.Size*4)
	for idx := 0; idx < cg.Base.Size; idx++ {
		res = append(res, cg.Get(node.NP(idx, 0)))
		res = append(res, cg.Get(node.NP(idx, cg.Base.Size-1)))
	}
	for idy := 0; idy < cg.Base.Size; idy++ {
		res = append(res, cg.Get(node.NP(0, idy)))
		res = append(res, cg.Get(node.NP(cg.Base.Size-1, idy)))
	}
	return res
}

//AccessibilityGrid generate an accessiblity grid from the compacted version of the grid. (wont alter current Base)
func (cg *CompoundedGrid) AccessibilityGrid() (res AccessibilityGridStruct) {
	g := Create(cg.Base.Size, nodetype.Plain)
	for idx, n := range cg.Base.Nodes {
		g.Nodes[idx] = n
	}
	for idx, n := range cg.Delta.Nodes {
		if n.Type != nodetype.None {
			g.Nodes[idx] = n
		}
	}

	return g.DefaultAccessibilityGrid()
}
