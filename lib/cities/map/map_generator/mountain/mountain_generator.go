package mountain_generator

import (
	"math"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

//MountainGenerator generate mountains ahah
type MountainGenerator struct {
	Width     tools.IntRange
	Range     tools.IntRange
	Disparity int
}

//Create a new mountain generator with randomized conf
func Create() *MountainGenerator {
	mg := new(MountainGenerator)
	mg.Width = tools.MakeIntRange(3, tools.RandInt(4, 6))
	mg.Range = tools.MakeIntRange(3, tools.RandInt(5, 15))
	mg.Disparity = 1
	return mg
}

//Level of the sub generator see Generator Level
func (mg *MountainGenerator) Level() map_generator.GeneratorLevel {
	return map_generator.Ground
}

//Generate Will apply generator to provided grid
func (mg *MountainGenerator) Generate(gd *grid.CompoundedGrid) error {

	width := mg.Width.Roll()
	rg := mg.Range.Roll()

	pt := tools.MakeIntRange(0, gd.Base.Size-1)

	test := 0
	// test 3 times to get the right place for a nice mountain, failure ? don't care ... :)
	for test < 3 {
		nd := node.NP(pt.Roll(), pt.Roll())
		done := false
		if !gd.IsFilled(nd) {
			targets := node.PointsAtDistance(nd, rg, gd.Base.Size)
			lentarget := len(targets)
			for i := 0; i < lentarget; i++ {
				target := targets[tools.RandInt(0, lentarget-1)]
				if gd.IsFilled(target) {

					div := math.Sqrt(math.Pow(float64(target.X-nd.X), 2) + math.Pow(float64(target.Y-nd.Y), 2))

					// unit vector = { X/V(X²+Y²), Y/V(X²+Y²) }
					unitX := int(float64(target.X-nd.X) / div)
					unitY := int(float64(target.Y-nd.Y) / div)

					for idx := width - mg.Disparity; idx < (rg - (width - mg.Disparity)); idx = idx + width + mg.Disparity {
						for _, nd := range node.PointsWithinInDistance(node.NP(unitX*idx, unitY*idx), width, gd.Base.Size) {
							gd.SetP(nd.X, nd.Y, node.Mountain)
						}
					}

					done = true
					break
				}
			}
		}
		if done {
			break
		}
		test++
	}

	// tried three times to add a mountain range, but couldn't ... that's okay.
	return nil
}

//Name of the generator
func (mg *MountainGenerator) Name() string {
	return "MountainGenerator"
}
