package pattern

import (
	"math"
	"upsilon_cities_go/lib/cities/node"
)

//Pattern pattern of points to be extracted from a grid
type Pattern []node.Point

//Adjascent pattern
var Adjascent = Pattern{node.NP(-1, 0), node.NP(0, 1), node.NP(1, 0), node.NP(0, -1)}

//Square pattern
var Square = Pattern{node.NP(-1, 0), node.NP(-1, 1), node.NP(0, 1), node.NP(1, 1), node.NP(1, 0), node.NP(1, -1), node.NP(0, -1), node.NP(-1, -1)}

var circlePatterns = make(map[int]Pattern)

//Apply pattern to provided point, limited by map size
func (p Pattern) Apply(loc node.Point, mapSize int) (res []node.Point) {
	for _, v := range p {
		n := loc.Add(v)
		if n.IsIn(mapSize) {
			res = append(res, n)
		}
	}
	return res
}

//ApplyBorders pattern to provided point, limited by map size
func (p Pattern) ApplyBorders(loc node.Point, mapSize int) (res []node.Point) {
	for _, v := range p {
		n := loc.Add(v)
		if n.IsAdjBorder(mapSize) {
			res = append(res, n)
		}
	}
	return res
}

//GenerateCirclePattern generate a circle pattern, if necessary.
func GenerateCirclePattern(size int) (res Pattern) {
	if val, ok := circlePatterns[size]; ok {
		return val
	}

	len := 0
	for i := 0.0; i < 2*math.Pi; i += 0.2 {
		s, c := math.Sincos(i)
		np := node.NP(((int)(c * (float64)(size))), ((int)(s * (float64)(size))))
		if len > 0 {
			last := res[len-1]
			if node.Distance(last, np) != 0 {
				res = append(res, np)
				len++
			}
		} else {
			res = append(res, np)
			len++

		}
	}

	circlePatterns[size] = res
	return res
}

//GenerateLinePattern this one wont be stored ...
func GenerateLinePattern(to node.Point) (res Pattern) {

	dist := math.Sqrt(math.Pow(float64(to.X), 2) + math.Pow(float64(to.Y), 2))

	// unit vector = { X/V(X²+Y²), Y/V(X²+Y²) }
	unitX := float64(to.X) / dist
	unitY := float64(to.Y) / dist

	for idx := 0; idx < int(dist); idx++ {
		nd := node.NP(int(math.Round(unitX*float64(idx))), int(math.Round(unitY*float64(idx))))
		res = append(res, nd)
	}

	return res
}
