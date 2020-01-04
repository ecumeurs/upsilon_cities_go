package grid

import (
	"math"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

//FillSquare will add a square on the grid
func (gd *Grid) FillSquare(typ node.NodeType, dist int, center node.Point) {
	for x := tools.Max(0, center.X-dist); x < tools.Min(gd.Size, center.X+1+dist); x++ {
		for y := tools.Max(0, center.Y-dist); y < tools.Min(gd.Size, center.Y+1+dist); y++ {
			gd.GetP(x, y).Type = typ
		}
	}
}

//FillCircle will add a Circle on the grid
func (gd *Grid) FillCircle(typ node.NodeType, dist int, center node.Point) {
	for _, nd := range node.PointsWithinInCircle(center, dist, gd.Size) {
		gd.Get(nd).Type = typ
	}
}

//AddLine will add a Line on the grid
func (gd *Grid) AddLine(typ node.NodeType, from node.Point, to node.Point, width int) {

	dist := math.Sqrt(math.Pow(float64(to.X-from.X), 2) + math.Pow(float64(to.Y-from.Y), 2))

	// unit vector = { X/V(X²+Y²), Y/V(X²+Y²) }
	unitX := float64(to.X-from.X) / dist
	unitY := float64(to.Y-from.Y) / dist

	for idx := 0; idx < int(dist); idx++ {
		nd := node.NP(from.X+int(unitX*float64(idx)), from.Y+int(unitY*float64(idx)))
		gd.Get(nd).Type = typ
		// now apply Width

		locDist := math.Sqrt(math.Pow(float64(to.X-nd.X), 2) + math.Pow(float64(to.Y-nd.Y), 2))

		locUnitX := -(float64(to.Y-nd.Y) / locDist)
		locUnitY := float64(to.X-nd.X) / locDist
		for widx := -width; widx <= width; widx++ {
			nd := node.NP(nd.X+int(locUnitX*float64(widx)), nd.Y+int(locUnitY*float64(widx)))
			gd.Get(nd).Type = typ
		}
	}
}
