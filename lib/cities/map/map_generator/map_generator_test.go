package map_generator

import (
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/forest_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/mountain_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/sea_generator"
	"upsilon_cities_go/lib/cities/node"
)

func TestGenerateSimpleT1Map(t *testing.T) {

	base := grid.Create(20, node.Plain)

	mg := mountain_generator.Create()

	mapgen := New()
	mapgen.AddGenerator(mg)

	mapgen.Generate(base)
	t.Errorf(base.String())

}

func TestGenerateComplexT1Map(t *testing.T) {

	base := grid.Create(20, node.Plain)

	mg := mountain_generator.Create()
	sg := sea_generator.Create()

	mapgen := New()
	mapgen.AddGenerator(mg)
	mapgen.AddGenerator(sg)

	mapgen.Generate(base)
}

func TestGenerateSimpleT2Map(t *testing.T) {

	base := grid.Create(20, node.Plain)

	mg := mountain_generator.Create()
	fg := forest_generator.Create()

	mapgen := New()
	mapgen.AddGenerator(mg)
	mapgen.AddGenerator(mg)
	mapgen.AddGenerator(fg)

	mapgen.Generate(base)

	t.Errorf(base.String())
}
