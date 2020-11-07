package map_generator

import (
	"log"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/nodetype"
)

//MapSubGenerator build
type MapSubGenerator interface {
	// Level of the sub generator see Generator Level
	Level() map_level.GeneratorLevel
	// Will apply generator to provided grid
	Generate(grid *grid.CompoundedGrid) error
	// Name of the generator
	Name() string
}

//MapGenerator build a new grid
type MapGenerator struct {
	Size       int
	Generators map[map_level.GeneratorLevel][]MapSubGenerator
}

//New build a new mapgenerator fully initialized.
func New() (mg *MapGenerator) {
	mg = new(MapGenerator)
	mg.Size = 20
	mg.Generators = make(map[map_level.GeneratorLevel][]MapSubGenerator)
	return
}

//Generate will generate a new grid based on available generators and their respective configuration
func (mg MapGenerator) Generate() (*grid.Grid, error) {
	var cg grid.CompoundedGrid

	cg.Base = grid.Create(mg.Size, nodetype.Plain)
	cg.Delta = grid.Create(mg.Size, nodetype.None)

	for level, arr := range mg.Generators {
		cg.Delta = grid.Create(mg.Size, nodetype.None)

		for _, v := range arr {
			err := v.Generate(&cg)
			if err != nil {
				log.Fatalf("MapGenerator: Failed to apply Generator Lvl: %d %s", level, v.Name())
				return nil, err
			}
		}
		cg.Base = cg.Compact()
		cg.Delta = grid.Create(mg.Size, nodetype.None)
	}
	g := cg.Compact()
	return g, nil
}

//AddGenerator Add A generator to the stack
func (mg *MapGenerator) AddGenerator(gen MapSubGenerator) {
	mg.Generators[gen.Level()] = append(mg.Generators[gen.Level()], gen)
	return
}
