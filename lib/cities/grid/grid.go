package grid

import (
	"math/rand"
	"sort"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/generator"
)

//Grid content of map, note `json:"-"` means it won't be exported as json ...
type Grid struct {
	ID         int
	Nodes      []node.Node
	Name       string
	LastUpdate time.Time
	Cities     map[int]*city.City
	// Helper to get back to a city by it's pos.
	LocationToCity map[int]*city.City `json:"-"`
	Size           int
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

//New create a new random grid.
func New(dbh *db.Handler) *Grid {
	grid := new(Grid)
	grid.ID = 0
	grid.LastUpdate = time.Now()

	// generate map ... size

	grid.generate(dbh, 20, 3)

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

func evaluateCandidates(cty *city.City, candidates map[int]*city.City) (candidateNeigbours neighbours) {
	// seek nearest cities, discard cities where distance > 10
	var cn neighbours
	knownNeighbours := make(map[int]int)
	for _, v := range cty.Neighbours {
		knownNeighbours[v.ID] = v.ID
	}

	for _, candidate := range candidates {
		// can't have a too highly connected city ;)
		if cty.Location != candidate.Location {
			// exclude already neighbours ;) of course.
			if _, found := knownNeighbours[candidate.ID]; !found {
				ok, neighbour := evaluateCandidate(cty, candidate)
				if ok {
					cn = append(cn, *neighbour)
				}
			}
		}
	}

	// sort by min distance.
	sort.Slice(cn, func(i, j int) bool { return cn[i].Distance < cn[j].Distance })

	candidateNeigbours = cn
	// check containement
	for _, n := range cn {
		var ncandidates neighbours
		found := false
		for _, nn := range candidateNeigbours {
			if n.Cty.Location != nn.Cty.Location {
				similar, contained := nn.ProposedPath.Similar(n.ProposedPath, 2)
				if !(similar || contained) {
					ncandidates = append(ncandidates, nn)
				}
			} else {
				found = true
			}
		}
		if found {
			candidateNeigbours = append(ncandidates, n)
		}
	}

	return
}

//buildRoad will check all cities and build appropriate pathways
func (grid *Grid) buildRoad() {

	for _, cty := range grid.Cities {

		maxNeighbour := 3 + rand.Intn(3)
		// seek already bound neighbours

		maxNeighbour = maxNeighbour - len(cty.Neighbours)
		if maxNeighbour > 0 {

			// keep max

			newNeighbours := evaluateCandidates(cty, grid.Cities)[0:maxNeighbour]

			// from and to targeted cities

			knownNeighbours := make(map[int]int)

			for _, v := range cty.Neighbours {
				knownNeighbours[v.ID] = v.ID
			}

			for _, nei := range newNeighbours {
				if _, found := knownNeighbours[nei.Cty.ID]; found {
					continue
				}

				cty.Neighbours = append(cty.Neighbours, nei.Cty)
				nei.Cty.Neighbours = append(nei.Cty.Neighbours, cty)
				knownNeighbours[nei.Cty.ID] = nei.Cty.ID

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

//generate generate a new grid
func (grid *Grid) generate(dbh *db.Handler, maxSize int, scarcity int) {
	grid.Clear()
	grid.Size = maxSize
	currentID := 1
	currentCityID := -1 // use a negative id ... so that will be stored as new.
	var tmpCities []*city.City
	for i := 0; i < maxSize; i++ {

		for j := 0; j < maxSize; j++ {
			var nde node.Node
			nde.ID = currentID
			currentID++
			nde.Location.X = j
			nde.Location.Y = i
			nde.Type = grid.randomCity(nde.Location, scarcity)
			if nde.Type == node.CityNode {
				cty := new(city.City)
				cty.Name = generator.CityName()
				cty.Location = nde.Location
				cty.ID = currentCityID
				currentCityID--
				tmpCities = append(tmpCities, cty)
				grid.LocationToCity[nde.Location.Y*grid.Size+nde.Location.X] = cty
			}
			grid.Nodes = append(grid.Nodes, nde)
		}
	}

	grid.Insert(dbh)

	// how to handle neighbouring registration ...
	// city insert doesn't generate neighbours, but update will.
	// thus insert all cities then update them all !
	// not efficient but should be enough.
	for _, v := range tmpCities {
		v.Insert(dbh, grid.ID)
	}

	grid.Cities = make(map[int]*city.City)
	for _, v := range tmpCities {
		grid.Cities[v.ID] = v
	}

	grid.buildRoad()

	for _, v := range grid.Cities {
		v.Update(dbh)
	}

	grid.Update(dbh)
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

//randomCity assign a random city; the higher scarcity the lower the chance to have a city ;)
func (grid *Grid) randomCity(location node.Point, scarcity int) node.NodeType {
	roll := rand.Intn(scarcity + 1)
	if roll < scarcity {
		return node.None
	}

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
