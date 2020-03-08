package pattern

import (
	"math"
	"upsilon_cities_go/lib/cities/node"
)

//Pattern pattern of points to be extracted from a grid
type Pattern []node.Point

//Adjascent pattern
var Adjascent = Pattern{node.NP(-1, 0), node.NP(0, 1), node.NP(1, 0), node.NP(0, -1)}

var adjascentPatterns = make(map[int]Pattern)
var adjascentOutlinePatterns = make(map[int]Pattern)

//Square pattern
var Square = Pattern{node.NP(-1, 0), node.NP(-1, 1), node.NP(0, 1), node.NP(1, 1), node.NP(1, 0), node.NP(1, -1), node.NP(0, -1), node.NP(-1, -1)}

var squareOutlinePatterns = make(map[int]Pattern)

var circlePatterns = make(map[int]Pattern)

//Apply pattern to provided point, limited by map size
func (p Pattern) Apply(loc node.Point, mapSize int) []node.Point {
	res := make([]node.Point, 0, len(p))
	for idx := range p {
		n := loc.Add(p[idx])
		if n.IsIn(mapSize) {
			res = append(res, n)
		}
	}
	return res
}

//ApplyBorders pattern to provided point, limited by map size
func (p Pattern) ApplyBorders(loc node.Point, mapSize int) []node.Point {
	res := make([]node.Point, 0, len(p))
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
		np := node.NP(((int)(c * (float64)(size+1))), ((int)(s * (float64)(size+1))))
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

func makeAdjascent(toGenerate []node.Point, known *map[int]bool, dist int) (res []node.Point) {
	for _, n := range toGenerate {
		for _, w := range Adjascent {
			candidate := n.Add(w)
			candidateAbs := candidate.ToAbs(dist)
			_, found := (*known)[candidateAbs]
			if !found {
				(*known)[candidateAbs] = true
				res = append(res, candidate)
			}
		}
	}

	return
}

//GenerateAdjascentPattern Will produce ( or recover ) the pattern for adjascent items. default Adjascent pattern only provide 1dist .
func GenerateAdjascentPattern(dist int) (res Pattern) {
	if v, found := adjascentPatterns[dist]; found {
		return v
	}
	current := make([]node.Point, 0)
	current = append(current, node.NP(0, 0))
	res = append(res, node.NP(0, 0))
	known := make(map[int]bool)
	known[0] = true // that's starting node.

	for i := 0; i < dist; i++ {
		round := makeAdjascent(current, &known, dist)

		res = append(res, round...)
		current = round
	}
	adjascentPatterns[dist] = res
	return
}

//GenerateSquareOutlinePattern Will produce ( or recover ) the pattern for items in square dist.
func GenerateSquareOutlinePattern(dist int) (res Pattern) {
	if v, found := squareOutlinePatterns[dist]; found {
		return v
	}

	if dist == 0 {
		res = append(res, node.NP(0, 0))
		squareOutlinePatterns[dist] = res
		return
	}

	for x := -dist; x <= dist; x++ {
		for y := -dist; y <= dist; y++ {
			if x == -dist || y == -dist || x == dist || y == dist {
				res = append(res, node.NP(x, y))
			}
		}
	}

	squareOutlinePatterns[dist] = res

	return
}

//GenerateAdjascentOutlinePattern Will produce ( or recover ) the pattern for adjascent outline items.
func GenerateAdjascentOutlinePattern(dist int) (res Pattern) {
	if v, found := adjascentOutlinePatterns[dist]; found {
		return v
	}
	current := make([]node.Point, 0)
	current = append(current, node.NP(0, 0))
	res = append(res, node.NP(0, 0))
	known := make(map[int]bool)
	known[0] = true // that's starting node.

	for i := 0; i < dist; i++ {
		round := makeAdjascent(current, &known, dist)

		current = round
	}
	adjascentOutlinePatterns[dist] = current
	res = current
	return
}
