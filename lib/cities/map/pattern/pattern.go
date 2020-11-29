package pattern

import (
	"math"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

//Pattern pattern of points to be extracted from a grid
type Pattern []node.Point

//Adjascent pattern
var Adjascent = Pattern{node.NP(-1, 0), node.NP(0, 1), node.NP(1, 0), node.NP(0, -1)}

var adjascentPatterns = make(map[int]Pattern)
var adjascentOutlinePatterns = make(map[int]Pattern)
var adjascentOutlineWidthPatterns = make(map[int]map[int]Pattern)

//Square pattern
var Square = Pattern{node.NP(-1, 0), node.NP(-1, 1), node.NP(0, 1), node.NP(1, 1), node.NP(1, 0), node.NP(1, -1), node.NP(0, -1), node.NP(-1, -1)}

var squarePatterns = make(map[int]Pattern)
var squareOutlinePatterns = make(map[int]Pattern)
var squareOutlineWidthPatterns = make(map[int]map[int]Pattern)

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

//MakeAbsAdjascent generate a new array of points at are adjascent to provided array.
func MakeAbsAdjascent(toGenerate []node.Point, known *map[int]bool, dist int) (res Pattern) {
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

//MakeAdjascent generate a new array of points at are adjascent to provided array.
func MakeAdjascent(toGenerate []node.Point, known *map[int]bool, dist int) (res Pattern) {
	for _, n := range toGenerate {
		for _, w := range Adjascent {
			candidate := n.Add(w)
			candidateAbs := candidate.ToInt(dist)
			_, found := (*known)[candidateAbs]
			if !found {
				(*known)[candidateAbs] = true
				res = append(res, candidate)
			}
		}
	}

	return
}

func sign(p1, p2, p3 node.Point) float64 {
	return float64(p1.X-p3.X)*float64(p2.Y-p3.Y) - float64(p2.X-p3.X)*float64(p1.Y-p3.Y)
}

func pointInTriangle(pt, v1, v2, v3 node.Point) bool {
	var d1, d2, d3 float64
	var has_neg, has_pos bool

	d1 = sign(pt, v1, v2)
	d2 = sign(pt, v2, v3)
	d3 = sign(pt, v3, v1)

	has_neg = (d1 < 0) || (d2 < 0) || (d3 < 0)
	has_pos = (d1 > 0) || (d2 > 0) || (d3 > 0)

	return !(has_neg && has_pos)
}

//GenerateTriangle will generate a triangle to target destination
func GenerateTriangle(to node.Point, mapSize int, endWidth int) (res Pattern) {

	dist := math.Sqrt(math.Pow(float64(to.X), 2) + math.Pow(float64(to.Y), 2))

	// unit vector = { X/V(X²+Y²), Y/V(X²+Y²) }
	unitX := float64(to.X) / dist
	unitY := float64(to.Y) / dist

	alreadyIn := make(map[int]bool)

	t1 := node.NP(0, 0)
	t2 := to.Add(node.NP(int(math.Round(float64(endWidth/2)*unitY)), int(math.Round(float64(endWidth/2)*-unitX))))
	t3 := to.Add(node.NP(int(math.Round(float64(endWidth/2)*-unitY)), int(math.Round(float64(endWidth/2)*unitX))))

	minX := tools.Min(t1.X, tools.Min(t2.X, t3.X))
	maxX := tools.Max(t1.X, tools.Max(t2.X, t3.X))
	minY := tools.Min(t1.Y, tools.Min(t2.Y, t3.Y))
	maxY := tools.Max(t1.Y, tools.Max(t2.Y, t3.Y))

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			nd := node.NP(x, y)
			if !alreadyIn[nd.ToAbs(mapSize)] {
				if pointInTriangle(nd, t1, t2, t3) {
					alreadyIn[nd.ToAbs(mapSize)] = true
					res = append(res, nd)
				}
			}
		}
	}

	return res
}

//GenerateAdjascentPattern Will produce ( or recover ) the pattern for adjascent items. default Adjascent pattern only provide 1dist .
func GenerateAdjascentPattern(dist int) (res Pattern) {
	if v, found := adjascentPatterns[dist]; found {
		return v
	}

	if dist == 0 {
		res := append(make([]node.Point, 0), node.NP(0, 0))
		adjascentPatterns[dist] = res
		return res
	}
	if dist == 1 {
		res := Adjascent
		adjascentPatterns[dist] = res
		return res
	}

	current := make([]node.Point, 0)
	current = append(current, node.NP(0, 0))
	res = append(res, node.NP(0, 0))
	known := make(map[int]bool)
	known[node.NP(0, 0).ToAbs(dist)] = true

	for i := 0; i < dist; i++ {
		round := MakeAbsAdjascent(current, &known, dist)

		res = append(res, round...)
		current = round
	}
	adjascentPatterns[dist] = res
	return
}

//GenerateSquarePattern Will produce ( or recover ) the pattern for items in square dist.
func GenerateSquarePattern(dist int) (res Pattern) {
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

			res = append(res, node.NP(x, y))
		}
	}

	squarePatterns[dist] = res

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

//GenerateSquareOutlineWidthPattern Will produce ( or recover ) the pattern for square outline items and several before as well (according to width) note width can't be >= dist ;) if width == dist then use adj.
func GenerateSquareOutlineWidthPattern(dist int, width int) (res Pattern) {
	if _, f := adjascentOutlineWidthPatterns[dist]; !f {
		adjascentOutlineWidthPatterns[dist] = make(map[int]Pattern)
	} else {
		if v, found := adjascentOutlineWidthPatterns[dist][width]; found {
			return v
		}
	}

	if dist-width > 0 {

		current := GenerateSquareOutlinePattern(dist)

		if dist-width > 1 {
			for i := 1; i < width; i++ {
				add := GenerateSquareOutlinePattern(dist - i)
				current = append(current, add...)
			}
		}

		adjascentOutlineWidthPatterns[dist][width] = current
	}
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
	known[node.NP(0, 0).ToAbs(dist)] = true // that's starting node.

	for i := 0; i < dist; i++ {
		round := MakeAbsAdjascent(current, &known, dist)

		current = round
	}
	adjascentOutlinePatterns[dist] = current
	res = current
	return
}

//GenerateAdjascentOutlineWidthPattern Will produce ( or recover ) the pattern for adjascent outline items and several before as well (according to width) note width can't be >= dist ;) if width == dist then use adj.
func GenerateAdjascentOutlineWidthPattern(dist int, width int) (res Pattern) {
	if _, f := adjascentOutlineWidthPatterns[dist]; !f {
		adjascentOutlineWidthPatterns[dist] = make(map[int]Pattern)
	} else {
		if v, found := adjascentOutlineWidthPatterns[dist][width]; found {
			return v
		}
	}

	if dist-width > 1 {

		current := GenerateAdjascentPattern(dist)
		sub := GenerateAdjascentPattern(dist - width)

		adjascentOutlineWidthPatterns[dist][width] = node.Remove(current, sub)
		return adjascentOutlineWidthPatterns[dist][width]
	} else {
		if dist-width == 1 {

			adjascentOutlineWidthPatterns[dist][width] = GenerateAdjascentOutlinePattern(dist)
			return adjascentOutlineWidthPatterns[dist][width]
		}
	}
	return
}

//Outline generate a new array of points at are adjascent to provided array.
func Outline(toGenerate []node.Point, size int) (res []node.Point) {
	known := make(map[int]bool)
	for _, v := range toGenerate {
		known[v.ToInt(size)] = true
	}

	for _, n := range toGenerate {
		for _, w := range Adjascent {
			candidate := n.Add(w)
			candidateAbs := candidate.ToInt(size)
			if !known[candidateAbs] {
				known[candidateAbs] = true
				res = append(res, candidate)
			}
		}
	}

	return
}
