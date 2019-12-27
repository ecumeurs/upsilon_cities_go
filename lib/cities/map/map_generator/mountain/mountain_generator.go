package mountain_generator

import (
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator"
	"upsilon_cities_go/lib/cities/tools"
)

//MountainGenerator generate mountains ahah
type MountainGenerator struct {
	Width tools.IntRange
	Range tools.IntRange
}

//Create a new mountain generator with randomized conf
func Create() *MountainGenerator {
	mg := new(MountainGenerator)
	mg.Width = tools.MakeIntRange{3, tools.RandInt(4, 6)}
	mg.Range = tools.MakeIntRange{3, tools.RandInt(5, 15)}
	return mg
}

//Level of the sub generator see Generator Level
func (mg *MountainGenerator) Level() map_generator.GeneratorLevel {
	return map_generator.Ground
}

//Generate Will apply generator to provided grid
func (mg *MountainGenerator) Generate(grid *grid.Grid) error {

	return nil
}

//Name of the generator
func (mg *MountainGenerator) Name() string {
	return "MountainGenerator"
}
