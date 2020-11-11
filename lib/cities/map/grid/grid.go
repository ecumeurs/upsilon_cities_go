package grid

import (
	"log"
	"math/rand"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

//State used by grid evolution
type State struct {
	NextCaravan   time.Time
	NextCaravanID int
}

//Grid content of map, note `json:"-"` means it won't be exported as json ...
//Note This is the main holder for most items of a Map ;)
type Grid struct {
	ID         int
	Nodes      []node.Node
	Name       string
	LastUpdate time.Time
	Cities     map[int]*city.City
	Size       int
	Base       nodetype.GroundType

	// Helpers
	LocationToCity map[int]*city.City `json:"-"`
	Evolution      State              `json:"-"`
}

//ShortGrid only provide most basic of informations (for index stuff)
type ShortGrid struct {
	ID         int
	Name       string
	LastUpdate time.Time
}

//Clear a grid
func (grid *Grid) Clear() {
	grid.Nodes = make([]node.Node, 0)
	grid.Cities = make(map[int]*city.City)
	grid.LocationToCity = make(map[int]*city.City)
}

//Create a new grid based on requested size.
func Create(size int, base nodetype.GroundType) *Grid {
	gd := new(Grid)
	gd.Size = size

	gd.Nodes = make([]node.Node, 0, size*size)
	gd.Cities = make(map[int]*city.City)
	gd.LocationToCity = make(map[int]*city.City)

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			n := node.New(j, i)
			n.ID = i*size + j
			n.Ground = base
			gd.Nodes = append(gd.Nodes, n)
		}
	}

	return gd
}

//String stringify
func (grid *Grid) String() string {
	var res string
	i := 0
	res = "\n"
	for _, node := range grid.Nodes {
		res += node.Short() + " "
		i++
		if i == grid.Size {
			res += "\n"
			i = 0
		}
	}
	return res
}

//GetCityByLocation will get a city matching a location.
func (grid *Grid) GetCityByLocation(location node.Point) *city.City {
	if location.X < 0 || location.X >= grid.Size {
		return nil
	}
	if location.Y < 0 || location.Y >= grid.Size {
		return nil
	}

	return grid.LocationToCity[location.Y*grid.Size+location.X]
}

//Store grid in database for the first time. includes everything necessary
func Store(dbh *db.Handler, gd *Grid) {
	// generate appropriate number of corporations ...

	nbCorporations := len(gd.Cities)/3 + 1
	corps := make(map[int]*corporation.Corporation)
	toSet := make([]*corporation.Corporation, 0, nbCorporations)

	for i := 0; i < nbCorporations; i++ {
		corp := corporation.New(gd.ID)
		corp.Insert(dbh)
		corps[corp.ID] = corp
		toSet = append(toSet, corp)
	}

	// assign corporations to cities ...

	unused := assignCorps(gd.Cities, toSet)

	for _, v := range gd.Cities {
		v.Update(dbh)
	}

	// drop unused corporations ...
	for _, v := range unused {
		v.Drop(dbh)
	}

	gd.Update(dbh)
}

func assignNeighboursCorp(neighbours []*city.City, cities map[int]*city.City, corp *corporation.Corporation, nb int, citiesAssigned []*city.City) (bool, []*city.City) {
	if nb == 0 {
		return true, citiesAssigned
	}
	if len(neighbours) == 0 {
		return false, citiesAssigned
	}

	cty := neighbours[0]

	if cty.CorporationID == 0 {
		cty.Fame[corp.ID] = 500
		cty.CorporationID = corp.ID
		corp.CitiesID = append(corp.CitiesID, cty.ID)

		neighbours = neighbours[1:]

		log.Printf("Grid: Sub Assigning corp %d to city %d ", corp.ID, cty.ID)
		for _, v := range cty.NeighboursID {
			n := cities[v]
			if n.CorporationID == 0 {
				neighbours = append(neighbours, n)
			}
		}
		citiesAssigned = append(citiesAssigned, cty)

		return assignNeighboursCorp(neighbours, cities, corp, nb-1, citiesAssigned)
	}

	return assignNeighboursCorp(neighbours[1:], cities, corp, nb, citiesAssigned)
}

func assignCorps(cities map[int]*city.City, toSet []*corporation.Corporation) []*corporation.Corporation {
	if len(toSet) == 0 {
		return toSet
	}

	curCorp := toSet[0]

	for _, v := range cities {
		// seek a city without corps ... assume they'll all have enough neighbours anyway.
		if v.CorporationID == 0 {
			v.CorporationID = curCorp.ID
			v.Fame[curCorp.ID] = 500
			curCorp.CitiesID = append(curCorp.CitiesID, v.ID)

			neighbours := make([]*city.City, 0)
			for _, w := range v.NeighboursID {
				n := cities[w]
				if n.CorporationID == 0 {
					neighbours = append(neighbours, n)
				}
			}

			citiesAssigned := make([]*city.City, 0)
			citiesAssigned = append(citiesAssigned, v)

			okay, citiesAssigned := assignNeighboursCorp(neighbours, cities, curCorp, 2, citiesAssigned)
			if !okay {
				for _, w := range citiesAssigned {
					w.CorporationID = 0
					delete(w.Fame, curCorp.ID)
				}
				// try with another city
				// Means this city will be a singleton. Singleton are handled at the end of the recursive by the late check
				// see below ;)

				delete(v.Fame, curCorp.ID)
				continue
			}

			return assignCorps(cities, toSet[1:])
		}
	}

	reusedCorps := make(map[int]bool)
	// check for singleton
	for k, v := range cities {
		if v.CorporationID == 0 {
			// link it with another corp group.
			for _, w := range v.NeighboursID {
				n := cities[w]
				if n.CorporationID != 0 && !reusedCorps[n.CorporationID] {
					v.CorporationID = n.CorporationID
					reusedCorps[n.CorporationID] = true
					cities[k] = v
					v.Fame[v.CorporationID] = 500
					break
				}

			}
		} else {
			v.Fame[v.CorporationID] = 500
		}
	}

	// no city without corporation found.
	return toSet
}

//Get will seek out a node.
func (grid *Grid) Get(location node.Point) *node.Node {
	if location.X > grid.Size-1 {
		log.Fatalf("Grid: Get: X %d out of bound %d", location.X, grid.Size)
		return nil
	}
	if location.Y > grid.Size-1 {
		log.Fatalf("Grid: Get: Y %d out of bound %d", location.Y, grid.Size)
		return nil
	}
	if grid.Size*location.Y+location.X >= len(grid.Nodes) {
		log.Fatalf("Grid: Get: Location %d out of bound %d", grid.Size*location.Y+location.X, len(grid.Nodes))
		return nil
	}
	return &grid.Nodes[grid.Size*location.Y+location.X]
}

//GetP will seek out a node.
func (grid *Grid) GetP(x int, y int) *node.Node {
	if !tools.InEq(x, 0, grid.Size-1) {
		log.Fatalf("Grid: GetP: X %d out of bound %d", x, grid.Size)
		return nil
	}
	if !tools.InEq(y, 0, grid.Size-1) {
		log.Fatalf("Grid: GetP: Y %d out of bound %d", y, grid.Size)
		return nil
	}
	if grid.Size*y+x >= len(grid.Nodes) {
		log.Fatalf("Grid: GetP: Location %d out of bound %d", grid.Size*y+x, len(grid.Nodes))

		return nil
	}
	return &grid.Nodes[grid.Size*y+x]
}

//GetRange fetch nodes in range.
func (grid *Grid) GetRange(location node.Point, reach int) []*node.Node {
	location.X = location.X - reach/2
	location.Y = location.Y - reach/2

	var res []*node.Node

	for i := 0; i < reach; i++ {
		for j := 0; j < reach; j++ {
			pt := grid.GetP(location.X+j, location.Y+i)
			if pt != nil {
				res = append(res, pt)
			}
		}
	}
	return res
}

//GetAtRange fetch nodes at range(circle).
func (grid *Grid) GetAtRange(location node.Point, reach int) []*node.Node {
	pts := node.PointsAtDistance(location, reach, grid.Size)
	res := make([]*node.Node, 0, len(pts))

	for _, p := range pts {
		res = append(res, grid.Get(p))
	}

	return res
}

//randomCity assign a random city; the higher scarcity the lower the chance to have a city ;)
func (grid *Grid) randomCity(location node.Point, scarcity int) bool {
	roll := rand.Intn(scarcity + 1)
	if roll < scarcity {
		return false
	}

	// seek target location and a nice square of 3
	// if no cities are present in there then try it

	interloppers := grid.GetRange(location, 6)
	for _, nd := range interloppers {
		if nd.IsStructure {
			return false
		}
	}

	return false
}

//SelectPattern will select corresponding nodes in a grid based on pattern & location
func (grid *Grid) SelectPattern(loc node.Point, pattern pattern.Pattern) []*node.Node {
	res := make([]*node.Node, 0, len(pattern))
	for _, v := range pattern.Apply(loc, grid.Size) {
		res = append(res, grid.Get(v))
	}
	return res
}

//SelectPatternIf will select corresponding nodes in a grid based on pattern & location if match predicate
func (grid *Grid) SelectPatternIf(loc node.Point, pattern pattern.Pattern, predicate func(node.Node) bool) []*node.Node {
	res := make([]*node.Node, 0, len(pattern))
	for _, v := range pattern.Apply(loc, grid.Size) {
		if predicate(*grid.Get(v)) {
			res = append(res, grid.Get(v))
		}
	}
	return res
}
