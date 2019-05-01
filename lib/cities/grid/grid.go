package grid

import (
	"math/rand"
	"sort"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

type Grid struct {
	ID         int
	Nodes      []node.Node
	Name       string
	LastUpdate time.Time
	Cities     []*city.City
	// Helper to get back to a city by it's pos.
	LocationToCity map[int]*city.City
	Size           int
}

//Clear a grid
func (grid *Grid) Clear() {
	grid.Nodes = nil
	grid.Cities = nil
	grid.LocationToCity = make(map[int]*city.City)
}

//New create a new random grid.
func New() *Grid {
	grid := new(Grid)
	grid.ID = 0
	grid.LastUpdate = time.Now()

	// generate map ... size

	grid.Generate(20, 3)
	// grid has been generated randomly ... now clear out unwanted cities (those not matching)
	grid.BuildRoad()

	return grid
}

//String stringify
func (grid *Grid) String() string {
	var res string
	i := 0
	for _, node := range grid.Nodes {
		res += node.Short()
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

//neighbour is a helper struct for build road. its a simple link to a city with a distance.
type neighbour struct {
	Distance     int
	Cty          *city.City
	ProposedPath node.Path
}

type neighbours []neighbour

// check wether cities contains target
func containsCity(cities []*city.City, target *city.City) bool {
	for _, v := range cities {
		if v.Location == target.Location {
			return true
		}
	}
	return false
}

func evaluateCandidate(cty *city.City, candidate *city.City) (ok bool, nei *neighbour) {
	ok = false
	nei = nil
	if len(candidate.Neighbours) < 5 {
		// well obviously it would be stupid to add it if its already a neighbour
		if !containsCity(cty.Neighbours, candidate) {

			npath := node.MakePath(cty.Location, candidate.Location)

			nei = new(neighbour)
			nei.Distance = node.Distance(cty.Location, candidate.Location)
			nei.Cty = candidate
			nei.ProposedPath = npath
			ok = true
		}
	}
	return
}

func evaluateCandidates(cty *city.City, candidates []*city.City) (candidateNeigbours neighbours) {
	// seek nearest cities, discard cities where distance > 10
	for _, candidate := range candidates {
		// can't have a too highly connected city ;)
		if cty.Location != candidate.Location {
			ok, neighbour := evaluateCandidate(cty, candidate)
			if ok {
				candidateNeigbours = append(candidateNeigbours, *neighbour)
			}
		}
	}

	// sort by min distance.
	sort.Slice(candidateNeigbours, func(i, j int) bool { return candidateNeigbours[i].Distance < candidateNeigbours[j].Distance })

	return
}

//BuildRoad will check all cities and build appropriate pathways
func (grid *Grid) BuildRoad() {

	for _, cty := range grid.Cities {

		maxNeighbour := 3 + rand.Intn(3)
		// seek already bound neighbours

		maxNeighbour = maxNeighbour - len(cty.Neighbours)
		if maxNeighbour > 0 {

			// keep max

			newNeighbours := evaluateCandidates(cty, grid.Cities)[0:maxNeighbour]

			// from and to targeted cities

			for _, nei := range newNeighbours {
				cty.Neighbours = append(cty.Neighbours, nei.Cty)
				nei.Cty.Neighbours = append(nei.Cty.Neighbours, cty)

				// build pathway
				var toPathway node.Pathway
				toPathway.FromCityID = cty.ID
				toPathway.ToCityID = nei.Cty.ID
				toPathway.Road = nei.ProposedPath

				cty.Roads = append(cty.Roads, toPathway)

				var fromPathway node.Pathway
				fromPathway.FromCityID = nei.Cty.ID
				fromPathway.ToCityID = cty.ID
				fromPathway.Road = make([]node.Point, len(nei.ProposedPath), len(nei.ProposedPath))

				for i := 0; i < len(nei.ProposedPath); i++ {

					step := nei.ProposedPath[len(nei.ProposedPath)-(i+1)]
					fromPathway.Road[i] = step
					// by the way mark them as road as well ...

					if i != 0 && i != (len(nei.ProposedPath)-1) {
						grid.Nodes[step.ToInt(grid.Size)].Type = node.Road
					}
				}

				nei.Cty.Roads = append(nei.Cty.Roads, fromPathway)
			}
		}
	}
}

//Generate generate a new grid
func (grid *Grid) Generate(maxSize int, scarcity int) {
	grid.Clear()
	grid.Size = maxSize
	currentID := 0
	currentCityID := 0
	for i := 0; i < maxSize; i++ {

		for j := 0; j < maxSize; j++ {
			var nde node.Node
			nde.ID = currentID
			currentID++
			nde.Location.X = j
			nde.Location.Y = i
			nde.Type = grid.RandomCity(nde.Location, scarcity)
			if nde.Type == node.CityNode {
				cty := new(city.City)
				cty.Location = nde.Location
				cty.ID = currentCityID
				currentCityID++
				grid.Cities = append(grid.Cities, cty)
				grid.LocationToCity[nde.Location.Y*grid.Size+nde.Location.X] = cty
			}
			grid.Nodes = append(grid.Nodes, nde)
		}
	}
}

//Get will seek out a node.
func (grid *Grid) Get(location node.Point) *node.Node {
	if location.X > grid.Size-1 {
		return nil
	}
	if location.Y > grid.Size-1 {
		return nil
	}
	if grid.Size*location.Y+location.X >= len(grid.Nodes) {
		return nil
	}
	return &grid.Nodes[grid.Size*location.Y+location.X]
}

//GetP will seek out a node.
func (grid *Grid) GetP(x int, y int) *node.Node {
	if !tools.InEq(x, 0, grid.Size-1) {
		return nil
	}
	if !tools.InEq(y, 0, grid.Size-1) {
		return nil
	}
	if grid.Size*y+x >= len(grid.Nodes) {
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

//RandomCity assign a random city; the higher scarcity the lower the chance to have a city ;)
func (grid *Grid) RandomCity(location node.Point, scarcity int) node.NodeType {
	roll := rand.Intn(scarcity + 1)
	if roll < scarcity {
		return node.None
	} else {
		// seek target location and a nice square of 3
		// if no cities are present in there then try it

		interloppers := grid.GetRange(location, 6)
		for _, nd := range interloppers {
			if nd.Type == node.CityNode {
				return node.None
			}
		}

		return node.CityNode

	}
}

//IsValid check grid validity
func (grid *Grid) IsValid() bool {
	return true
}
