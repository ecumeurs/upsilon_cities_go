package river_generator

import (
	"log"
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
	// x5: a meander measure at least 5 cells.
	directness := mg.Directness.Roll() * 5

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
							// same border check -- refuse same border for origin and target ...
							if nde.X == 0 && candidate.Location.X == 0 {
								continue
							}
							if nde.X == gd.Base.Size-1 && candidate.Location.X == gd.Base.Size-1 {
								continue
							}
							if nde.Y == 0 && candidate.Location.Y == 0 {
								continue
							}
							if nde.Y == gd.Base.Size-1 && candidate.Location.Y == gd.Base.Size-1 {
								continue
							}

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
		retry := false

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

		//log.Printf("RiverGenerator: Selected an origin: %v -> Target: %v", origin.String(), target.String())
		//log.Printf("RiverGenerator: Distance: %d, real %f", node.Distance(origin, target), node.RealDistance(origin, target))
		//log.Printf("RiverGenerator: Targeted distance: %d", length)

		// generate a AStar based on this.

		{
			var current = make(map[int]node.Point)
			current[target.ToInt(gd.Base.Size)] = target
			var next = make(map[int]node.Point)

			currentDist := 0

			for _, v := range tempGrid.AvailableCells {
				tempGrid.SetData(v, -1)
			}

			for true {
				for _, v := range current {
					if tempGrid.GetData(v) == -1 {
						tempGrid.SetData(v, currentDist)
						for _, w := range tempGrid.SelectPattern(v, pattern.Adjascent) {
							if !gd.IsFilled(w.Location) && tempGrid.GetData(w.Location) == -1 {
								if _, ok := next[w.Location.ToInt(gd.Base.Size)]; !ok {
									next[w.Location.ToInt(gd.Base.Size)] = w.Location
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
		}

		/*
			{
				currentY := 0
				for _, v := range tempGrid.Nodes {
					if v.Location.Y != currentY {
						fmt.Printf("\n")
						currentY = v.Location.Y
					}
					fmt.Printf("%2d:%2d %03d ", v.Location.X, v.Location.Y, tempGrid.GetData(v.Location))
				}
				fmt.Printf("\n")
			}
		*/
		// AStar completed !
		river := make([]node.Point, 0)
		{
			river = append(river, origin)

			current := origin
			previous := origin
			targetLength := node.Distance(target, origin) + directness
			currentScore := tempGrid.GetData(origin)
			used := make(map[int]bool)

			//log.Printf("RiverGenerator: Total Distance: %d", targetLength)

			for true {

				//log.Printf("RiverGenerator: Current Cell: %v Current Score: %d", current.String(), currentScore)
				foundOne := false

				candidates := tempGrid.SelectPatternIf(current, pattern.Adjascent, func(n node.Node) bool {
					if _, ok := used[n.Location.ToInt(gd.Base.Size)]; !ok {
						return !gd.IsFilled(n.Location)
					}
					return false
				})

				n := len(candidates)
				for i := 0; i < len(candidates); i++ {
					randIndex := rand.Intn(n)
					candidates[n-1], candidates[randIndex] = candidates[randIndex], candidates[n-1]
				}

				//log.Printf("RiverGenerator: len candidates: %d", len(candidates))

				randreach := 1.0 - ((float32)(targetLength-currentScore) / (float32)(targetLength))

				if rand.Float32() > randreach {
					for _, v := range candidates {

						score := tempGrid.GetData(v.Location)

						//log.Printf("RiverGenerator: Candidate: %v Score %d randreach %f ", v.Location.String(), score, randreach)

						if currentScore > score {
							continue
						}

						if targetLength > currentScore && score <= targetLength {
							targetLength--

							previous = current
							current = v.Location
							currentScore = score
							foundOne = true
							river = append(river, v.Location)
							used[v.Location.ToInt(gd.Base.Size)] = true

							break // No need to continue
						}

					}
					if !foundOne {
						// seek smallest
						smallest := current
						smallestScore := 99999
						for _, v := range candidates {
							score := tempGrid.GetData(v.Location)
							//log.Printf("RiverGenerator: Candidate: %v Score %d  ", v.Location.String(), score)

							if smallestScore > score {
								smallest = v.Location
								smallestScore = score
							}
						}

						targetLength--
						previous = current
						current = smallest
						currentScore = smallestScore
						foundOne = true
						river = append(river, smallest)
						used[smallest.ToInt(gd.Base.Size)] = true
					}
				} else {
					for _, v := range candidates {
						score := tempGrid.GetData(v.Location)
						//log.Printf("RiverGenerator: Candidate: %v Score %d  ", v.Location.String(), score)

						if currentScore > score {
							targetLength--
							previous = current
							current = v.Location
							currentScore = score
							foundOne = true
							river = append(river, v.Location)
							used[v.Location.ToInt(gd.Base.Size)] = true
							break // No need to continue
						}
					}
					if !foundOne {

						smallest := current
						smallestScore := 99999
						for _, v := range candidates {
							score := tempGrid.GetData(v.Location)
							// seek smallest
							if smallestScore > score {
								smallest = v.Location
								smallestScore = score
							}

						}

						targetLength--
						previous = current
						current = smallest
						currentScore = smallestScore
						foundOne = true
						river = append(river, smallest)
						used[smallest.ToInt(gd.Base.Size)] = true
					}
				}

				if !foundOne {
					log.Printf("RiverGenerator: Unable to find a next cell")
					// that's super weird, it means that we didn't find any acceptable next node in current stuff.
					retry = true
					break
				} else {
					tempGrid.Apply(previous, pattern.Adjascent, func(n *node.Node, dt int) (ndata int) {
						if dt != 0 {
							return dt + 3
						}
						return dt
					})
				}

				if currentScore == 0 {
					log.Printf("RiverGenerator: Reached end of path successfully: target length ? %d", targetLength)

					if targetLength < -3 || targetLength > 3 {
						// well, we're way out of expected bounds.
						retry = true
					}
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
