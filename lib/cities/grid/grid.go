package grid

import (
	"log"
	"math/rand"
	"sort"
	"time"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/generator"
)

type gridEvolution struct {
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

	// Helpers
	LocationToCity map[int]*city.City `json:"-"`
	Evolution      gridEvolution      `json:"-"`
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
	AlreadyIn    bool
}

type neighbours []neighbour

// check wether cities contains target
func containsCity(cities []int, target int) bool {
	for _, v := range cities {
		if target == v {
			return true
		}
	}
	return false
}

func evaluateCandidate(cty *city.City, candidate *city.City) (ok bool, nei neighbour) {
	ok = false
	if len(candidate.NeighboursID) < 5 {
		// well obviously it would be stupid to add it if its already a neighbour
		if !containsCity(cty.NeighboursID, candidate.ID) {

			npath := node.MakePath(cty.Location, candidate.Location)
			nei.Distance = node.Distance(cty.Location, candidate.Location)
			nei.Cty = candidate
			nei.ProposedPath = npath
			nei.AlreadyIn = false
			ok = true
		}
	}
	return
}

func evaluateCandidates(cty *city.City, candidates map[int]*city.City) (candidateNeigbours neighbours) {
	// seek nearest cities, discard cities where distance > 10
	var cn neighbours
	knownNeighbours := make(map[int]int)
	for _, v := range cty.NeighboursID {
		knownNeighbours[v] = v
	}

	for _, candidate := range candidates {
		// can't have a too highly connected city ;)
		if cty.Location != candidate.Location {
			// exclude already neighbours ;) of course.
			if _, found := knownNeighbours[candidate.ID]; !found {
				ok, neighbour := evaluateCandidate(cty, candidate)
				if ok {
					cn = append(cn, neighbour)
				}
			} else {
				// add them for path checker.
				var neighbour neighbour
				neighbour.Cty = candidate
				neighbour.Distance = node.Distance(cty.Location, candidate.Location)
				neighbour.ProposedPath = node.MakePath(cty.Location, candidate.Location)
				neighbour.AlreadyIn = true
				cn = append(cn, neighbour)
			}
		}
	}

	// sort by min distance.
	sort.Slice(cn, func(i, j int) bool { return cn[i].Distance < cn[j].Distance })

	candidateNeigbours = cn

	log.Printf("Grid: Checking neighbours of city: %s", cty.Name)
	var ncandidates neighbours

	// keep only distance < 10
	for _, n := range cn {
		if n.Distance > 10 {
			continue
		}
		ncandidates = append(ncandidates, n)
	}

	candidateNeigbours = ncandidates
	cn = ncandidates

	rejected := make(map[int]int)

	// check containement
	for _, n := range cn {
		if _, found := rejected[n.Cty.ID]; found {
			continue
		}

		log.Printf("Grid: Checking neighbouring city: %s distance %d", n.Cty.Name, n.Distance)

		var ncandidates neighbours
		for _, nn := range candidateNeigbours {

			if nn.AlreadyIn {
				if n.Cty.Location != nn.Cty.Location {
					log.Printf("Grid: Keeps %s because already in", nn.Cty.Name)
					ncandidates = append(ncandidates, nn)
				}
				continue
			}

			if n.Cty.Location != nn.Cty.Location {
				similar, _, contains := nn.ProposedPath.Similar(n.ProposedPath, 2)
				if similar {
					log.Printf("Grid: Rejects: %s (%d) because Similar", n.Cty.Name, nn.Distance)
					rejected[n.Cty.ID] = n.Cty.ID
					ncandidates = append(ncandidates, nn)
					break
				} else if !contains {
					ncandidates = append(ncandidates, nn)
				} else {
					log.Printf("Grid: Rejects: %s (%d) because contains", nn.Cty.Name, nn.Distance)
					rejected[nn.Cty.ID] = nn.Cty.ID
				}
			}
		}

		if _, found := rejected[n.Cty.ID]; found {
			continue
		}

		candidateNeigbours = append(ncandidates, n)
	}

	return
}

//buildRoad will check all cities and build appropriate pathways
func (grid *Grid) buildRoad() {

	for _, cty := range grid.Cities {

		maxNeighbour := 3 + rand.Intn(3)
		// seek already bound neighbours

		maxNeighbour = maxNeighbour - len(cty.NeighboursID)
		if maxNeighbour > 0 {

			// keep max

			newNeighbours := evaluateCandidates(cty, grid.Cities)
			knownNeighbours := make(map[int]int)
			for _, v := range cty.NeighboursID {
				knownNeighbours[v] = v
			}

			nb := make([]string, 0)
			for _, v := range newNeighbours {
				nb = append(nb, v.Cty.Name)
			}
			log.Printf("Grid: Selected neighbours %v keep %d", nb, maxNeighbour)
			if len(newNeighbours) == 0 {
				continue
			} else {
				var nc neighbours
				for _, v := range newNeighbours {
					if _, found := knownNeighbours[v.Cty.ID]; !found {
						nc = append(nc, v)
						if len(nc) == maxNeighbour {
							break
						}
					}
				}
				newNeighbours = nc
			}

			nb = make([]string, 0)
			for _, v := range newNeighbours {
				nb = append(nb, v.Cty.Name)
			}
			log.Printf("Grid: Selected neighbours %v", nb)

			for _, nei := range newNeighbours {
				if _, found := knownNeighbours[nei.Cty.ID]; found {
					continue
				}

				cty.NeighboursID = append(cty.NeighboursID, nei.Cty.ID)
				nei.Cty.NeighboursID = append(nei.Cty.NeighboursID, cty.ID)
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
	grid.Name = generator.RegionName()
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
				cty := city.New()
				cty.Name = generator.CityName()
				cty.Location = nde.Location
				cty.Storage.SetSize(300)
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

	// generate appropriate number of corporations ...

	nbCorporations := len(grid.Cities)/3 + 1
	corps := make(map[int]*corporation.Corporation)
	toSet := make([]*corporation.Corporation, 0)

	for i := 0; i < nbCorporations; i++ {
		corp := corporation.New(grid.ID)
		corp.Insert(dbh)
		corps[corp.ID] = corp
		toSet = append(toSet, corp)
	}

	// assign corporations to cities ...

	unused := assignCorps(grid.Cities, toSet)

	for _, v := range grid.Cities {
		v.Update(dbh)
	}

	// drop unused corporations ...
	for _, v := range unused {
		v.Drop(dbh)
	}

	grid.Update(dbh)
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
