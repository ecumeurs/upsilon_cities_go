package map_generator

import (
	"log"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/db"
)

//MapSubGenerator build
type MapSubGenerator interface {
	// Level of the sub generator see Generator Level
	Level() map_level.GeneratorLevel
	// Will apply generator to provided grid
	Generate(grid *grid.CompoundedGrid, dbh *db.Handler) error
	// Name of the generator
	Name() string
}

//MapGenerator build a new grid
type MapGenerator struct {
	Size       int
	Base       nodetype.NodeType
	Generators map[map_level.GeneratorLevel][]MapSubGenerator
}

//New build a new mapgenerator fully initialized.
func New() (mg *MapGenerator) {
	mg = new(MapGenerator)
	mg.Size = 20
	mg.Base = nodetype.Plain
	mg.Generators = make(map[map_level.GeneratorLevel][]MapSubGenerator)
	return
}

//Generate will generate a new grid based on available generators and their respective configuration
func (mg MapGenerator) Generate(dbh *db.Handler) (*grid.Grid, error) {
	var cg grid.CompoundedGrid

	cg.Base = grid.Create(mg.Size, mg.Base)
	cg.Base.Insert(dbh) // ensure we get an ID !

	cg.Delta = grid.Create(mg.Size, nodetype.None)

	for level, arr := range mg.Generators {
		cg.Delta = grid.Create(mg.Size, nodetype.None)

		for _, v := range arr {
			err := v.Generate(&cg, dbh)
			if err != nil {
				log.Fatalf("MapGenerator: Failed to apply Generator Lvl: %d %s", level, v.Name())
				return nil, err
			}
		}
		cg.Base = cg.Compact()
	}
	g := cg.Compact()
	g.Update(dbh)
	return g, nil
}

//AddGenerator Add A generator to the stack
func (mg *MapGenerator) AddGenerator(gen MapSubGenerator) {
	mg.Generators[gen.Level()] = append(mg.Generators[gen.Level()], gen)
	return
}
