package map_generator

import (
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/forest_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/mountain_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/sea_generator"
	"upsilon_cities_go/lib/cities/node"
)

// TestGenerateSimpleT1Map create a simple map 20x20 with nothing else but a mountain.
func TestGenerateSimpleT0Map(t *testing.T) {

	base := grid.Create(20, node.Plain)

	mg := mountain_generator.Create()

	mapgen := New()
	mapgen.AddGenerator(mg)

	mapgen.Generate(base)
}

// This one allow multiples T1 obstacle to be found.
// Note: T0 obstacles are direct obstacle that limits what can be used on later tiers.
func TestGenerateComplexT0Map(t *testing.T) {

	base := grid.Create(20, node.Plain)

	mg := mountain_generator.Create()
	sg := sea_generator.Create()

	mapgen := New()
	mapgen.AddGenerator(mg)
	mapgen.AddGenerator(sg)

	mapgen.Generate(base)
}

// T1 means rivers.
// A river either goes from a mountain to a sea
// or from one border to the sea
// or from a mountain to a border
// or from a border to another.
func TestGenerateSimpleT1Map(t *testing.T) {

	base := grid.Create(20, node.Plain)

	mg := mountain_generator.Create()
	sg := sea_generator.Create()
	rg := river_generator.Create()

	mapgen := New()
	mapgen.AddGenerator(mg)
	mapgen.AddGenerator(sg)
	mapgen.AddGenerator(rg)

	mapgen.Generate(base)
}

// Forest can't be used on T0-1 stuff, so in this simple Test, mountains ranges shouldn't be cropped by forests.
func TestGenerateSimpleT2Map(t *testing.T) {

	base := grid.Create(20, node.Plain)

	mg := mountain_generator.Create()
	fg := forest_generator.Create()

	mapgen := New()
	mapgen.AddGenerator(mg)
	mapgen.AddGenerator(mg)
	mapgen.AddGenerator(fg)

	mapgen.Generate(base)
}
