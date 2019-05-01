package node

import (
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
}

//Short node type in short.
func (node *Node) Short() string {
	return node.Type.Short()
}

//ToInt convert a point in Arrayable int
func (loc Point) ToInt(size int) int {
	return loc.Y*size + loc.X
}

//Distance manhattan between two points.
func Distance(lhs, rhs Point) int {
	return tools.Abs(rhs.X-lhs.X) + tools.Abs(rhs.Y-lhs.Y)
}

//Similar tell whether a pathway contains another, with at most deviation
func (path *Path) Similar(other *Path, deviation int) (similar bool, totallyIncluded bool) {
	deviated := 0

	for _, nde := range *path {
		for _, onde := range *other {
			if onde != nde {
				deviated++
			}
		}
	}

	similar = (deviated + tools.Min(0, len(*other)-len(*path))) < deviation
	totallyIncluded = deviated == 0
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
