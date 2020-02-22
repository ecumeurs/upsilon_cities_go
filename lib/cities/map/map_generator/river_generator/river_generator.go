package river_generator

import (
	"math/rand"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

//RiverGenerator generate rivers ahah
type RiverGenerator struct {
	Directness tools.IntRange // tell how direct the river should go: 0 will use the most direct trajectory.
	Length     tools.IntRange // length of the river
}

//Create a new mountain generator with randomized conf
func Create() (mg RiverGenerator) {
	mg.Directness = tools.MakeIntRange(3, tools.RandInt(3, 10))
	mg.Length = tools.MakeIntRange(6, tools.RandInt(10, 20))
	return
}

//Level of the sub generator see Generator Level
func (mg RiverGenerator) Level() map_level.GeneratorLevel {
	return map_level.River
}

//Generate Will apply generator to provided grid
func (mg RiverGenerator) Generate(gd *grid.CompoundedGrid) error {

	// a river goes:
	// * from a moutain
	// * from a border
	// to:
	// * sea
	// * border

	length := mg.Length.Roll()
	directness := mg.Directness.Roll()

	// assuredly this one is needed.
	path := make(map[int]int)

	{
		originCandidates := make(map[int]node.Point, 0)

		// just to remember which we already checked.
		// supposedly origin ... may not be needed ;)
		origin := make(map[int]node.Point, 0)
		target := make(map[int]node.Point, 0)

		// Step 1: find mountain ranges nodes => cells with a mountain next to a plain cell.
		for _, nde := range gd.Base.Nodes {
			if nde.Type != node.Mountain {
				continue
			}
			originCandidates[nde.Location.ToInt(gd.Base.Size)] = nde.Location
		}

		for _, nde := range gd.SelectMapBorders() {
			if nde.Type != node.Plain {
				continue
			}
			originCandidates[nde.Location.ToInt(gd.Base.Size)] = nde.Location
		}

		for _, nde := range originCandidates {

			for _, v := range gd.SelectPattern(nde, pattern.Adjascent) {
				if gd.Get(v.Location).Type == node.Plain {
					// that's a candidate !!
					// Step 2: find sea ranges nodes

					candidates := make([]node.Node, 0)

					for _, issea := range gd.SelectPattern(nde, pattern.GenerateCirclePattern(length)) {
						if gd.Get(issea.Location).Type == node.Sea {
							// candidate !
							candidates = append(candidates, issea)
						}
					}
					candidates = append(candidates, gd.SelectPatternMapBorders(nde, pattern.GenerateCirclePattern(length))...)

					for _, candidate := range candidates {
						obstacleFound := false
						for _, isobstacle := range pattern.GenerateLinePattern(candidate.Location.Sub(nde)).Apply(nde, gd.Base.Size) {
							if gd.IsFilled(isobstacle) {
								obstacleFound = true
								break
							}
						}
						if !obstacleFound {
							// match !
							if _, ok := origin[nde.ToInt(gd.Base.Size)]; !ok {
								origin[nde.ToInt(gd.Base.Size)] = nde
							}
							if _, ok := target[candidate.Location.ToInt(gd.Base.Size)]; !ok {
								target[candidate.Location.ToInt(gd.Base.Size)] = candidate.Location
							}

							path[nde.ToInt(gd.Base.Size)] = candidate.Location.ToInt(gd.Base.Size)
						}
					}

				}
			}
		}

		if len(path) == 0 {
			// just no options using moutains to sea ...

			return nil
		}
	}

	// select a random couple origin -> target

	tempGrid := gd.AccessibilityGrid()

	tries := 3
	for tries > 0 {
		tries--

		var origin node.Point
		var target node.Point

		t := rand.Intn(len(path))
		for k, v := range path {
			if t == 0 {
				// that's the one !
				origin = node.FromInt(k, gd.Base.Size)
				target = node.FromInt(v, gd.Base.Size)
				delete(path, k)
				break
			}
			t--
		}

		// generate a AStar based on this.

		{
			var current = make([]node.Point, 0)
			current = append(current, target)
			var next = make([]node.Point, 0)

			currentDist := 1
			tempGrid.SetData(target, 0)

			for _, v := range tempGrid.AvailableCells {
				tempGrid.SetData(v, -1)
			}

			for true {
				for _, v := range current {
					tempGrid.Apply(v, pattern.Adjascent, func(n *node.Node, data int) (ndata int) {
						ndata = data
						if data == -1 {
							ndata = currentDist
							for _, w := range pattern.Adjascent.Apply(n.Location, gd.Base.Size) {
								if tempGrid.IsAccessible(w) && tempGrid.GetData(w) == -1 {
									next = append(next, w)
								}
							}
						}
						return
					})
				}
				current = next
				currentDist++
				next = make([]node.Point, 0)
				if len(current) == 0 {
					break
				}
			}
		}

		// AStar completed !

		// initiate river !
		n := gd.Get(origin)
		n.Type = node.River
		gd.Set(n)

		river := make([]node.Point, 0)
		retry := false
		{
			river = append(river, origin)

			current := origin
			targetLength := length + directness
			currentScore := tempGrid.GetData(origin)
			used := make(map[int]bool)

			for true {

				foundOne := false
				for _, v := range tempGrid.SelectPatternIf(current, pattern.Adjascent, func(n node.Node) bool {
					if _, ok := used[n.Location.ToInt(gd.Base.Size)]; !ok {
						return n.Type == node.Plain
					}
					return false
				}) {

					score := tempGrid.GetData(v.Location)

					if currentScore > score {
						targetLength--
						current = v.Location
						currentScore = score
						foundOne = true
						river = append(river, v.Location)
						break // No need to continue
					}

					if targetLength > currentScore && score <= targetLength && rand.Float32() > (1.0-(float32)((targetLength-currentScore)/targetLength)) {
						targetLength--
						current = v.Location
						currentScore = score
						foundOne = true
						river = append(river, v.Location)
						break // No need to continue
					}
				}

				if !foundOne {
					// that's super weird, it means that we didn't find any acceptable next node in current stuff.
					retry = true
					break
				}

				if currentScore == 0 {
					break
				}
			}

		}

		if retry {
			// reselect a path and try again
			continue
		}

		for _, v := range river {
			n := gd.Get(v)
			n.Type = node.River
			gd.Set(n)
		}

		break // success
	}

	return nil
}

//Name of the generator
func (mg RiverGenerator) Name() string {
	return "RiverGenerator"
}
