package map_generator

import (
	"log"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/node"
)

//GeneratorLevel tell what is overridable and what's not.
//Means what's set on a level can't be removed by other levels, with some exceptions ...
type GeneratorLevel int

const (
	Ground         GeneratorLevel = 0 // Sea, Mountains
	River          GeneratorLevel = 1 // River rolls from mountains to seas
	Landscape      GeneratorLevel = 2 // this is what we may find elsewhere ( forest, desert )
	Resource       GeneratorLevel = 3 // Ressource assignation: Note this is mostly macro assignation of ressource (like here are minerals, here are plants, with exceptions )
	Structure      GeneratorLevel = 4 // Structures, like Cities, may be set a bit anywhere ... with some exceptions.
	Transportation GeneratorLevel = 5 // Transportation level ( roads, mostly ) will only be applied between cities
)

//MapSubGenerator build
type MapSubGenerator interface {
	// Level of the sub generator see Generator Level
	Level() GeneratorLevel
	// Will apply generator to provided grid
	Generate(grid *grid.CompoundedGrid) error
	// Name of the generator
	Name() string
}

//MapGenerator build a new grid
type MapGenerator struct {
	Generators map[GeneratorLevel][]MapSubGenerator
}

//Generate will generate a new grid based on available generators and their respective configuration
func (mg MapGenerator) Generate(g *grid.Grid) error {
	var cg grid.CompoundedGrid
	cg.Base = g

	for level, arr := range mg.Generators {
		cg.Delta = grid.Create(g.Size, node.None)

		for _, v := range arr {
			err := v.Generate(&cg)
			if err != nil {
				log.Fatalf("MapGenerator: Failed to apply Generator Lvl: %d %s", level, v.Name())
				return err
			}
		}
		cg.Base = cg.Compact()
	}
	return nil
}

//AddGenerator Add A generator to the stack
func (mg *MapGenerator) AddGenerator(gen MapSubGenerator) {
	mg.Generators[gen.Level()] = append(mg.Generators[gen.Level()], gen)
	return
}
