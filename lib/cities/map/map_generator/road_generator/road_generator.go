package road_generator

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
)

const (
	north int = 0
	east  int = 1
	south int = 2
	west  int = 3
)

//RoadGenerator generate roads ahah
type RoadGenerator struct {
	RoadInfluenceCost int

	DefaultDepthReach int
	DefaultDepthCost  int

	Neighbours tools.IntRange

	Connection      map[int]bool
	CostDepth       map[nodetype.GroundType]int
	CostReach       map[nodetype.LandscapeType]int
	LTCostDepth     map[nodetype.LandscapeType]int
	LTCostFunctions map[nodetype.LandscapeType]func(node.Node, grid.AccessibilityGridStruct)
	CostFunctions   map[nodetype.GroundType]func(node.Node, grid.AccessibilityGridStruct)

	Roads []generatedRoad
}

//Name of the generator
func (rg RoadGenerator) Name() string {
	return "RoadGenerator"
}

//Create a new road generator with randomized conf
func Create() (rg RoadGenerator) {
	rg.Roads = make([]generatedRoad, 0)

	rg.RoadInfluenceCost = -3

	rg.DefaultDepthReach = 3
	rg.DefaultDepthCost = 2

	rg.Neighbours = tools.MakeIntRange(2, 4)

	rg.CostDepth = make(map[nodetype.GroundType]int)
	rg.CostDepth[nodetype.NoGround] = 0
	rg.CostDepth[nodetype.Desert] = 2
	rg.CostDepth[nodetype.Plain] = 1
	rg.CostDepth[nodetype.Sea] = 999

	rg.LTCostDepth = make(map[nodetype.LandscapeType]int)
	rg.LTCostDepth[nodetype.NoLandscape] = 0
	rg.LTCostDepth[nodetype.Forest] = 2
	rg.LTCostDepth[nodetype.Mountain] = 2
	rg.LTCostDepth[nodetype.River] = 15

	rg.CostReach = make(map[nodetype.LandscapeType]int)
	rg.CostReach[nodetype.Forest] = 1
	rg.CostReach[nodetype.Mountain] = 2
	rg.CostReach[nodetype.River] = 0

	rg.Connection = make(map[int]bool)
	rg.Connection[north] = true
	rg.Connection[east] = true
	rg.Connection[south] = true
	rg.Connection[west] = true

	rg.CostFunctions = make(map[nodetype.GroundType]func(node.Node, grid.AccessibilityGridStruct))
	rg.CostFunctions[nodetype.Desert] = rg.computeDefaultCost
	rg.CostFunctions[nodetype.Plain] = rg.computeDefaultCost
	rg.CostFunctions[nodetype.Sea] = rg.refuse

	rg.LTCostFunctions = make(map[nodetype.LandscapeType]func(node.Node, grid.AccessibilityGridStruct))
	rg.LTCostFunctions[nodetype.Forest] = rg.computeDefaultReachCost
	rg.LTCostFunctions[nodetype.River] = rg.computeRiverCost
	rg.LTCostFunctions[nodetype.Mountain] = rg.computeDefaultReachCost

	return
}

func (rg RoadGenerator) fetchLTCost(nt nodetype.LandscapeType) int {
	c, has := rg.LTCostDepth[nt]
	if has {
		return c
	}
	return rg.DefaultDepthCost
}

func (rg RoadGenerator) fetchCost(nt nodetype.GroundType) int {
	c, has := rg.CostDepth[nt]
	if has {
		return c
	}
	return rg.DefaultDepthCost
}

func (rg RoadGenerator) fetchReach(nt nodetype.LandscapeType) int {
	c, has := rg.CostReach[nt]
	if has {
		return c
	}
	return rg.DefaultDepthReach
}

func (rg RoadGenerator) computeDefaultReachCost(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.Apply(n.Location, pattern.GenerateAdjascentPattern(rg.fetchReach(n.Landscape)), func(nn *node.Node, data int) (newData int) {
		return data + int(math.Floor(node.RealDistance(nn.Location, n.Location)/float64(rg.fetchReach(n.Landscape))*float64(rg.fetchLTCost(n.Landscape))))
	})
}

func (rg RoadGenerator) computeDefaultCost(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, acc.GetData(n.Location)+rg.fetchCost(n.Ground))
}

func (rg RoadGenerator) computeRiverCost(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, acc.GetData(n.Location)+rg.fetchLTCost(nodetype.River))
}

func (rg RoadGenerator) refuse(n node.Node, acc grid.AccessibilityGridStruct) {
	acc.SetData(n.Location, 999)
}

//Level of the sub generator see Generator Level
func (rg RoadGenerator) Level() map_level.GeneratorLevel {
	return map_level.Transportation
}

func (rg RoadGenerator) computeCost(nd node.Node, acc grid.AccessibilityGridStruct) {
	cost, has := rg.CostFunctions[nd.Ground]

	// roads have a slight bonus
	if nd.IsRoad {
		acc.SetData(nd.Location, acc.GetData(nd.Location)-15)
	}

	if acc.IsAccessible(nd.Location) {
		if has {
			// apply ground cost
			cost(nd, acc)
		}

		cost, has = rg.LTCostFunctions[nd.Landscape]
		if has {
			// apply landscape cost
			cost(nd, acc)
		}

		// noisify
		acc.SetData(nd.Location, acc.GetData(nd.Location)+tools.RandInt(-10, 10))

		// avoid roads on borders...
		if nd.Location.X == 0 || nd.Location.X == acc.Size-1 {
			acc.SetData(nd.Location, acc.GetData(nd.Location)+50)
		}
		if nd.Location.Y == 0 || nd.Location.Y == acc.Size-1 {
			acc.SetData(nd.Location, acc.GetData(nd.Location)+50)
		}
	} else {
		// inaccessible thus 999
		rg.refuse(nd, acc)
	}
}

func (rg RoadGenerator) astarGrid(gd *grid.CompoundedGrid, tempGrid *grid.AccessibilityGridStruct, origin, target node.Point) {
	var current = make(map[int]node.Point)
	current[target.ToInt(gd.Base.Size)] = target
	var next = make(map[int]node.Point)

	// ensure restrictions due to mountains and stuff ain't a major impediment to cross to reach a city
	currentDist := -5
	// ensure target city is heavily weighted as a goal
	tempGrid.SetData(target, -50)

	used := make(map[int]bool)

	// astaring
	for true {

		for _, v := range current {
			if !used[v.ToInt(gd.Base.Size)] {
				tempGrid.SetData(v, (currentDist*6)+tempGrid.GetData(v))
				used[v.ToInt(gd.Base.Size)] = true
				for _, w := range tempGrid.SelectPattern(v, pattern.Adjascent) {
					if !used[w.ToInt(gd.Base.Size)] {
						if _, ok := next[w.ToInt(gd.Base.Size)]; !ok {
							next[w.ToInt(gd.Base.Size)] = w
						}
					}
				}
			}
		}
		current = next
		currentDist++
		next = make(map[int]node.Point)
		if len(current) == 0 {
			break
		}
	}

	// ensure target city is heavily weighted as a goal
	tempGrid.SetData(target, -99)
	// ensure origin city is heavily weighted as a goal
	tempGrid.SetData(origin, -88)

}

type generatedRoad struct {
	Road   node.Path
	Nodes  map[int]bool
	Cities map[int]bool
}

//Generate Will apply generator to provided grid
func (rg RoadGenerator) Generate(gd *grid.CompoundedGrid, dbh *db.Handler) error {
	// prepare a road heat map.
	// rivers are a solid +10 in difficulty
	// mountains and forest stacks +2 for each depth of each
	// desert stacks +3 for each depth
	// Preexisting roads count for a solid - 15 in difficulty
	// Area near a road is easier to access.

	// Make a list of cities to generate roads from.
	// Add them trice, attempt to multiply connections from cities to cities.
	cities := gd.Base.Cities
	clist := make([]int, 0)
	for k := range cities {
		clist = append(clist, k)
	}

	rand.Shuffle(len(clist), func(i, j int) { clist[i], clist[j] = clist[j], clist[i] })

	targetCities := make([]int, 0)
	for k := range cities {
		targetCities = append(targetCities, k)
	}

	// randomize target cities
	rand.Shuffle(len(targetCities), func(i, j int) { targetCities[i], targetCities[j] = targetCities[j], targetCities[i] })

	for _, k := range clist {
		originCity := cities[k]

		targetCity := gd.Base.Cities[targetCities[tools.RandInt(0, len(targetCities)-1)]]
		if targetCity == nil {
			log.Fatalf("Weird nil target city")
		}
		for targetCity.ID == k {
			targetCity = gd.Base.Cities[targetCities[tools.RandInt(0, len(targetCities)-1)]]
			if targetCity == nil {
				log.Fatalf("Weird nil target city")
			}
		}

		err := rg.generateRoad(gd, originCity, targetCity)
		if err != nil {
			return err
		}
	}

	//err := rg.generateNeighbours(gd, dbh)
	//if err != nil {
	//	return err
	//}

	return nil
}

func printRoad(gd *grid.CompoundedGrid, acc *grid.AccessibilityGridStruct, road generatedRoad) {
	var res string
	i := 0
	res = "\n"
	for _, node := range acc.Nodes {
		hasColor := false
		if acc.GetData(node.Location) == -88 {
			hasColor = true
			res += "\033[43m"
		} else if acc.GetData(node.Location) == -99 {
			hasColor = true
			res += "\033[45m"
		} else if gd.Base.GetCityByLocation(node.Location) != nil {
			hasColor = true
			res += "\033[42m"
		} else if road.Nodes[node.Location.ToInt(acc.Size)] {
			hasColor = true
			res += "\033[41m"
		}
		res += fmt.Sprintf("%3d ", acc.GetData(node.Location))
		if hasColor {
			res += "\033[0m"
		}
		i++
		if i == acc.Size {
			res += "\n"
			i = 0
		}
	}
	log.Printf("RG: \n%s", res)
}
func printOtherRoad(gd *grid.CompoundedGrid, acc *grid.AccessibilityGridStruct, road generatedRoad, roadOther generatedRoad, endp node.Point) {
	var res string
	i := 0
	res = "\n"
	for _, node := range acc.Nodes {
		hasColor := false
		if acc.GetData(node.Location) == -88 {
			hasColor = true
			res += "\033[43m"
		} else if node.Location.IsEq(endp) && gd.Base.GetCityByLocation(node.Location) != nil {
			hasColor = true
			res += "\033[45;96m"
		} else if acc.GetData(node.Location) == -99 {
			hasColor = true
			res += "\033[45m"
		} else if node.Location.IsEq(endp) {
			hasColor = true
			res += "\033[96m"
		} else if gd.Base.GetCityByLocation(node.Location) != nil {
			hasColor = true
			res += "\033[42m"
		} else if road.Nodes[node.Location.ToInt(acc.Size)] {
			hasColor = true
			res += "\033[41m"
		} else if roadOther.Nodes[node.Location.ToInt(acc.Size)] {
			hasColor = true
			res += "\033[46m"
		}
		res += fmt.Sprintf("%3d ", acc.GetData(node.Location))
		if hasColor {
			res += "\033[0m"
		}
		i++
		if i == acc.Size {
			res += "\n"
			i = 0
		}
	}
	log.Printf("RG: \n%s", res)
}

//purgePrevious will check if current point is adj to a previous item in the road (except direct previous, of course)
// if that's the case, will remove all points inbetween
func purgePrevious(gd *grid.CompoundedGrid, gr *generatedRoad, currentLocation node.Point, acc *grid.AccessibilityGridStruct) {
	if len(gr.Road) < 2 {
		return // no need
	}

	log.Printf("RG: Purging road test")
	last := gr.Road[len(gr.Road)-2]
	
	for k := range gr.Nodes {
		gr.Nodes[k] = false
	}

	for _,v := range gr.Road {
		gr.Nodes[v.ToInt(acc.Size)] = true
	}

	for _, v := range pattern.Adjascent.Apply(currentLocation, acc.Size) {
		if last.IsEq(v) {
			continue
		}
		if !v.IsValid(acc.Size) {
			continue
		}

		if gr.Nodes[v.ToInt(acc.Size)] {
			// candidate !

			log.Printf("RG: Purging road current location %v shortcut %v", currentLocation, v)
			log.Printf("RG: Road %v", gr.Road)

			
			printRoad(gd, acc, *gr)
			
			// remove last point (it's itself.)
			gr.Nodes[gr.Road[len(gr.Road)-1].ToInt(acc.Size)] = false
			gr.Road = gr.Road[:len(gr.Road)-1]


			for len(gr.Road) > 1 && !gr.Road[len(gr.Road)-1].IsEq(v) {

				itm := gr.Road[len(gr.Road)-1]
				acc.SetData(itm, 999) // forcefully ignore tile.
				gr.Road = gr.Road[:len(gr.Road)-1]
				gr.Nodes[itm.ToInt(acc.Size)] = false
				log.Printf("RG: Road %v", gr.Road)
			}

			gr.Road = append(gr.Road, currentLocation)
			purgePrevious(gd, gr, currentLocation, acc)
			return
		}
	}
}

func (rg *RoadGenerator) generateRoad(gd *grid.CompoundedGrid, originCity *city.City, targetCity *city.City) error {
	log.Printf("RG: Generating road options %s -> %s", originCity.Location.String(), targetCity.Location.String())

	acc := gd.AccessibilityGrid()
	for x := 0; x < gd.Base.Size; x++ {
		for y := 0; y < gd.Base.Size; y++ {
			acc.SetData(node.NP(x, y), 0)
		}
	}

	// on cherche les villes :) (on les met dans un array ? :) )
	for x := 0; x < gd.Base.Size; x++ {
		for y := 0; y < gd.Base.Size; y++ {
			nd := gd.GetP(x, y)
			rg.computeCost(nd, acc)
		}
	}

	// on selection 2 villes (originCity, targetCity)
	rg.astarGrid(gd, &acc, originCity.Location, targetCity.Location)

	var gr generatedRoad
	gr.Cities = make(map[int]bool)
	gr.Nodes = make(map[int]bool)
	gr.Cities[originCity.ID] = true
	gr.Cities[targetCity.ID] = true
	gr.Road = make([]node.Point, 0)

	validRoadFound := false

	// on cherche le chemin le plus court et le moins cher
	currentLocation := originCity.Location
	target := targetCity.Location

	gr.Road = append(gr.Road, currentLocation)
	gr.Nodes[currentLocation.ToInt(gd.Base.Size)] = true

	for currentLocation != target {
		// seek lowest point from current location.

		// this is a potential candidate for a new road block
		if gd.Delta.Get(currentLocation).IsRoad {
			oneFound := false
			// seek out on which road existing road it is.
			for ridx, v := range rg.Roads {
				if v.Nodes[currentLocation.ToInt(gd.Base.Size)] {
					oneFound, validRoadFound, currentLocation = rg.roadFound(gd, &acc, originCity, targetCity, currentLocation, &gr, v, ridx)
					if validRoadFound || oneFound {
						break
					}
				}
			}
			if validRoadFound {
				break
			}
			if !oneFound {
				// go back
				if len(gr.Road) > 1 {
					acc.SetData(currentLocation, 999)
					gr.Nodes[currentLocation.ToInt(gd.Base.Size)] = false
					currentLocation = gr.Road[len(gr.Road)-1]
					gr.Road = gr.Road[:len(gr.Road)-1]
					continue // retry loop :)
				}
			}

			if currentLocation.IsEq(target) {
				break
			}
		}

		currentLowest := 999
		var currentPoint node.Point
		for _, targetNode := range acc.SelectPattern(currentLocation, pattern.Adjascent) {
			val := acc.GetData(targetNode)

			if targetNode.IsEq(target) {
				currentLowest = val
				currentPoint = targetNode
				break
			}

			if gr.Nodes[targetNode.ToInt(gd.Base.Size)] {
				// cant go backward
				continue
			}
			if val < currentLowest {
				currentLowest = val
				currentPoint = targetNode
			}
			if val == currentLowest && node.Distance(targetNode, target) < node.Distance(targetNode, target) {
				currentLowest = val
				currentPoint = targetNode
			}
		}

		if currentLowest >= 999 {
			if len(gr.Road) == 1 {
				printRoad(gd, &acc, gr)
				return fmt.Errorf("city has no road options %s -> %s", originCity.Location.String(), target.String())
			}
			// go backward and mark this one at 999
			// go backward if there are already 3 ajd road to this one.
			acc.SetData(currentLocation, 999)
			gr.Nodes[currentLocation.ToInt(gd.Base.Size)] = false
			currentLocation = gr.Road[len(gr.Road)-1]
			gr.Road = gr.Road[:len(gr.Road)-1]
			continue
		}
		gr.Road = append(gr.Road, currentPoint)
		gr.Nodes[currentPoint.ToInt(gd.Base.Size)] = true

		if gd.Base.Get(currentPoint).Landscape == nodetype.River {
			//ensure nearby rivers are out of scope, can't have 2 rivers roads...

			for _, targetNode := range acc.SelectPattern(currentLocation, pattern.Adjascent) {

				if gd.Base.Get(targetNode).Landscape == nodetype.River {
					// cant go backward
					acc.SetData(targetNode, 999)
				}
			}
		}

		currentLocation = currentPoint

		purgePrevious(gd, &gr, currentLocation, &acc)

	}
	// on stock qu'on a bien relier les deux villes au reseau.

	if !validRoadFound {
		rg.Roads = append(rg.Roads, gr)
		printRoad(gd, &acc, gr)
	}

	// on met le flag road sur toute les cases.

	for _, rloc := range gr.Road {
		gd.SetPRoad(rloc.X, rloc.Y, true)
	}
	log.Printf("Map: \n%s", gd.Delta.String())
	return nil
}

func (rg *RoadGenerator) roadFound(gd *grid.CompoundedGrid, acc *grid.AccessibilityGridStruct, originCity *city.City, targetCity *city.City, currentLocation node.Point, gr *generatedRoad, v generatedRoad, ridx int) (oneFound bool, validRoadFound bool, newCurrentLocation node.Point) {
	// it's linked to this road
	// maybe it'll be nicer to go on from anywhere else on this road.

	if v.Nodes[targetCity.Location.ToInt(gd.Base.Size)] {
		// winwinwinwin

		gr.Road = append(gr.Road, currentLocation)
		gr.Nodes[currentLocation.ToInt(gd.Base.Size)] = true

		v.Cities[originCity.ID] = true
		v.Cities[targetCity.ID] = true
		v.Road = append(v.Road, gr.Road...)
		for k, w := range gr.Nodes {
			v.Nodes[k] = w
		}
		rg.Roads[ridx] = v
		// merged current road with existing road.
		validRoadFound = true
		log.Printf("GR: joining another road and meet target city")
		printOtherRoad(gd, acc, *gr, v, targetCity.Location)
		return
	}

	refuse := make(map[int]bool)
	for k, v := range v.Nodes {
		refuse[k] = v
	}
	for k, v := range gr.Nodes {
		refuse[k] = v
	}
	currentLower := 999
	currentTarget := targetCity.Location

	// seek all adjascent tiles to road, refuse all road items .. ofcourse :)
	for _, v := range pattern.MakeAdjascent(v.Road, &refuse, gd.Base.Size) {
		if v.IsValid(gd.Base.Size) {
			if v.IsEq(targetCity.Location) {
				currentTarget = v
				currentLower = -99
				break 
			}
			score := node.Distance(currentLocation, v)*3 + node.Distance(v, targetCity.Location)*5 + acc.GetData(v)

			if score < currentLower {
				currentTarget = v
				currentLower = score
			}
		}
	}

	if currentLower == 999 {
		// must go backward and avoid that one at all cost.
		return false, false, currentLocation
	}

	log.Printf("GR: joining another road")
	printOtherRoad(gd, acc, *gr, v, currentTarget)

	oneFound = true
	for k, v := range v.Cities {
		gr.Cities[k] = v
	}
	gr.Road = append(gr.Road, v.Road...)
	for k, w := range v.Nodes {
		gr.Nodes[k] = w
	}

	gr.Road = append(gr.Road, currentLocation)
	gr.Nodes[currentLocation.ToInt(gd.Base.Size)] = true
	gr.Road = append(gr.Road, currentTarget)
	gr.Nodes[currentTarget.ToInt(gd.Base.Size)] = true

	// remove from old roads, will be reinserted later.
	rg.Roads = append(rg.Roads[:ridx], rg.Roads[ridx+1:]...)

	newCurrentLocation = currentTarget

	return
}

func (rg *RoadGenerator) generateNeighbours(gd *grid.CompoundedGrid, dbh *db.Handler) error {

	citiesLocations := make([]node.Point, len(gd.Base.Cities))
	for _, v := range gd.Base.Cities {
		citiesLocations = append(citiesLocations, v.Location)
		gd.SetPRoad(v.Location.X, v.Location.Y, true)
	}

	// find neighbours for each cities.
	for k, v := range gd.Base.Cities {
		targetNeighbours := rg.Neighbours.Roll()

		log.Printf("RG: city locations: %v", citiesLocations)
		//distNgb: citylocation -> dist
		distNgb, err := gd.Delta.RoadDistanceBetweenTargets(v.Location, citiesLocations)
		if err != nil {
			// not shouldn't error here :)
			log.Fatalf("RG: Shouldn't have errored here...: %s", err)
		}

		log.Printf("RG: Got distances from city %v to all cities: %v", v.Location, distNgb)

		rDistNgb := make(map[int]*city.City)
		orderedDist := make([]int, len(distNgb))
		for location, distance := range distNgb {
			cty := gd.Base.GetCityByLocation(node.FromInt(location, gd.Base.Size))
			if cty == nil {
				log.Fatalf("RG: Expected to find city at location: %d (%v) but got nil", location, node.FromInt(location, gd.Base.Size))
			}
			_, has := rDistNgb[distance]
			for has {
				distance++
				_, has = rDistNgb[distance]
			}
			rDistNgb[distance] = cty
			orderedDist = append(orderedDist, distance)
		}

		sort.Ints(orderedDist)

		testedNgb := make(map[int]bool)

		for _, w := range v.NeighboursID {
			testedNgb[w] = true
		}

		for _, w := range orderedDist {
			ncty := rDistNgb[w]

			if _, has := testedNgb[ncty.ID]; !has {
				if len(ncty.NeighboursID) < targetNeighbours {
					ncty.NeighboursID = append(ncty.NeighboursID, k)
					v.NeighboursID = append(v.NeighboursID, ncty.ID)
					ncty.Update(dbh)
				}
			}
			if len(v.NeighboursID) >= targetNeighbours {
				break
			}
		}

		v.Update(dbh)
	}
	return nil
}
