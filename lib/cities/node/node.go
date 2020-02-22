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
func PointsAtDistance(origin Point, distance int, mapSize int) (res []Point) {
	len := 0
	for i := 0.0; i < 2*math.Pi; i += 0.2 {
		s, c := math.Sincos(i)
		np := NP(origin.X+((int)(c*(float64)(distance))), origin.Y+((int)(s*(float64)(distance))))
		if len > 0 {
			if np.X < mapSize && np.Y < mapSize && np.X >= 0 && np.Y >= 0 {
				last := res[len-1]
				if Distance(last, np) != 0 {
					res = append(res, np)
					len++
				}
			}
		} else {
			if np.X < mapSize && np.Y < mapSize {
				res = append(res, np)
				len++
			}
		}
	}
	return res
}

//PointsWithinInDistance return all points within a distance
func PointsWithinInDistance(origin Point, distance int, size int) (res []Point) {
	for i := tools.Max(0, origin.X-distance); i <= origin.X+distance && i < size; i++ {
		for j := tools.Max(origin.Y-distance, 0); j <= origin.Y+distance && j < size; j++ {
			if tools.Abs(origin.X-i)+tools.Abs(origin.Y-j) <= distance {
				res = append(res, NP(i, j))
			}
		}
	}
	return res
}

//PointsWithinInCircle return all points within a distance
func PointsWithinInCircle(origin Point, distance int, size int) (res []Point) {
	for i := tools.Max(0, origin.X-distance); i <= origin.X+distance && i < size; i++ {
		for j := tools.Max(origin.Y-distance, 0); j <= origin.Y+distance && j < size; j++ {
			dist := math.Sqrt(math.Pow(float64(origin.X-i), 2) + math.Pow(float64(origin.Y-j), 2))

			if int(math.Round(dist)) <= distance {
				res = append(res, NP(i, j))
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

//FromInt convert int value to a point
func FromInt(value int, mapSize int) (res Point) {
	res.Y = value / mapSize
	res.X = value % mapSize
	return
}

//Add two points
func (loc Point) Add(n Point) (res Point) {
	res.X = n.X + loc.X
	res.Y = n.Y + loc.Y
	return
}

//Sub remove from n from loc
func (loc Point) Sub(n Point) (res Point) {
	res.X = loc.X - n.X
	res.Y = loc.Y - n.Y
	return
}

//IsIn check whether point is in mapsize.
func (loc Point) IsIn(mapSize int) bool {
	return loc.X >= 0 && loc.X < mapSize && loc.Y >= 0 && loc.Y < mapSize
}

//IsAdjBorder tell whether this point is within map BUT adj to a border.
func (loc Point) IsAdjBorder(mapSize int) bool {
	return loc.IsIn(mapSize) && (loc.X == 0 || loc.Y == 0 || loc.X == mapSize-1 || loc.Y == mapSize-1)
}

func (loc Point) String() string {
	return fmt.Sprintf("{%d,%d}", loc.X, loc.Y)
}

//Distance manhattan between two points.
func Distance(lhs, rhs Point) int {
	return tools.Abs(rhs.X-lhs.X) + tools.Abs(rhs.Y-lhs.Y)
}

//IsAdj tell whether rhs is adjacent strictly (no diag) to current point. Note IsAjd != same point ( if lhs==rhs => IsAdj = false )
func (lhs Point) IsAdj(rhs Point) bool {
	if rhs.X == lhs.X {
		return rhs.Y == lhs.Y-1 || rhs.Y == lhs.Y+1
	} else if rhs.Y == lhs.Y {
		return rhs.X == lhs.X-1 || rhs.X == lhs.X+1
	}
	return false
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
