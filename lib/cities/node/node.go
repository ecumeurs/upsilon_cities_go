package node

import (
	"fmt"
	"math"
	"upsilon_cities_go/lib/cities/tools"
)

type Point struct {
	X int
	Y int
}

type Path []Point

type Pathway struct {
	Road       Path
	FromCityID int
	ToCityID   int
}

type Node struct {
	ID       int
	Location Point
	Type     NodeType
	HasRoad  bool
}

//NP Create a new point
func NP(x, y int) (p Point) {
	p.X = x
	p.Y = y
	return
}

//New create a new node
func New(x, y int) (n Node) {
	n.Location = NP(x, y)
	n.HasRoad = false
	n.Type = None
	return n
}

//PointsAtDistance return all points at a distance
func PointsAtDistance(origin Point, distance int, size int) (res []Point) {
	len := 0
	for i := 0.0; i < 2*math.Pi; i += 0.2 {
		s, c := math.Sincos(i)
		np := NP(origin.X+((int)(c*(float64)(distance))), origin.Y+((int)(s*(float64)(distance))))
		if len > 0 {
			if np.X < size && np.Y < size && np.X >= 0 && np.Y >= 0 {
				last := res[len-1]
				if Distance(last, np) != 0 {
					res = append(res, np)
					len++
				}
			}
		} else {
			if np.X < size && np.Y < size {
				res = append(res, np)
				len++
			}
		}
	}
	return res
}

//PointsWithinInDistance return all points within a distance
func PointsWithinInDistance(origin Point, distance int, size int) (res []Point) {
	for i := 0; i < distance*2; i++ {
		if origin.X-distance+i > size {
			break
		}
		if origin.X-distance+i < 0 {
			continue
		}
		for j := 0; j < distance*2; j++ {
			if origin.Y-distance+j > size {
				break
			}
			if origin.Y-distance+j < 0 {
				continue
			}
			if i+j <= distance {
				res = append(res, NP(origin.X-distance+i, origin.Y-distance+j))
			}
		}
	}
	return res
}

//Short node type in short.
func (node *Node) Short() string {
	return node.Type.Short()
}

//ToInt convert a point in Arrayable int
func (loc Point) ToInt(size int) int {
	return loc.Y*size + loc.X
}

func (loc Point) String() string {
	return fmt.Sprintf("{%d,%d}", loc.X, loc.Y)
}

//Distance manhattan between two points.
func Distance(lhs, rhs Point) int {
	return tools.Abs(rhs.X-lhs.X) + tools.Abs(rhs.Y-lhs.Y)
}

//Contains check path contains point.
func (path Path) Contains(pt Point) bool {
	for _, v := range path {
		if v == pt {
			return true
		}
	}
	return false
}

//Similar tell whether a pathway contains another, with at most deviation
func (path Path) Similar(other Path, deviation int) (similar bool, totallyIncluded bool, includeOther bool) {
	deviated := 0
	totallyIncluded = true

	for _, nde := range path[1 : len(path)-1] {
		found := false
		for _, onde := range other[1 : len(other)-1] {
			if onde == nde {
				found = true
				break
			}
		}

		if !found || len(other[1:len(other)-1]) == 0 {
			deviated++
			totallyIncluded = false
		}
	}

	includeOther = true && len(other[1:len(other)-1]) > 0

	for _, nde := range other[1 : len(other)-1] {
		found := false
		for _, onde := range path[1 : len(path)-1] {
			if onde == nde {
				found = true
				break
			}
		}

		if !found || len(path[1:len(path)-1]) == 0 {
			includeOther = false
			break
		}
	}

	similar = (deviated + tools.Min(0, tools.Abs(len(other)-len(path)))) < deviation
	return
}

//MakePath create a path
func MakePath(from, to Point) Path {
	var res Path

	res = append(res, from)
	current := from

	for {
		if current == to {
			break
		}

		diffX := to.X - current.X
		diffY := to.Y - current.Y
		npoint := current
		if tools.Abs(diffX) < tools.Abs(diffY) {
			// work on Y
			if diffY < 0 {
				npoint.Y--
			} else {
				npoint.Y++
			}
			res = append(res, npoint)
		} else {
			// work on X
			if diffX < 0 {
				npoint.X--
			} else {
				npoint.X++
			}
			res = append(res, npoint)
		}
		current = npoint
	}

	return res
}
