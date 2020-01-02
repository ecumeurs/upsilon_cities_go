package city_generator

import (
	"log"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/tools"
)

//CityGenerator generate desert ahah
type CityGenerator struct {
	//Density number of cities in 10x10
	Density tools.IntRange
}

//Create a new desert generator with randomized conf
func Create() (mg CityGenerator) {
	mg.Density = tools.MakeIntRange(1, tools.RandInt(3, 5))
	return
}

//Level of the sub generator see Generator Level
func (mg CityGenerator) Level() map_level.GeneratorLevel {
	return map_level.Structure
}

//Generate Will apply generator to provided grid
func (mg CityGenerator) Generate(gd *grid.CompoundedGrid) error {
	density := mg.Density.Roll()
	size := gd.Base.Size
	nb := (size / 10) * density

	log.Printf("CityGenerator: Attempting to add Cities to map density: %d Size: %d => number of cities to add: %d", density, size, nb)

	// Note: a city may be added to the map
	// * in forest: <3 dist from plain
	// * in mountain: <1 dist from plain (adj to plain)
	// * in sea: <1 dist from plain 
	// * in desert <3 dist from plain 
	// cities may not be fully isolated ( completely surrounded by mountains, seas, deep forest ) 
	

	return nil
}

//Name of the generator
func (mg CityGenerator) Name() string {
	return "CityGenerator"
}
