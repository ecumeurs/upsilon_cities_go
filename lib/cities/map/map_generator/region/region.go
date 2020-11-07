package region

import (
	"errors"
	"upsilon_cities_go/lib/cities/map/map_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/city_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/forest_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/mountain_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/resource_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/river_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/road_generator"
	"upsilon_cities_go/lib/cities/map/map_generator/sea_generator"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
)

type generatorInclusion struct {
	Frequency int
	Generator map_generator.MapSubGenerator
}

type regionDefinition struct {
	AvailableGenerators []generatorInclusion
	ForcedGenerators    []map_generator.MapSubGenerator
	Usable              tools.IntRange
	Name                string
	Size                tools.IntRange
	Base                nodetype.NodeType
}

var regions map[string]regionDefinition

//Load initialize regions
func Load() {
	regions = make(map[string]regionDefinition)

	{
		var reg regionDefinition
		reg.Name = "Elvenwood"
		reg.Usable = tools.MakeIntRange(3, 5)
		reg.Size = tools.MakeIntRange(30, 50)
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{5, forest_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{2, mountain_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{1, river_generator.Create()})

		reg.ForcedGenerators = append(reg.ForcedGenerators, resource_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, city_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, road_generator.Create())

		regions[reg.Name] = reg
	}

	{
		var reg regionDefinition
		reg.Name = "Highlands"
		reg.Usable = tools.MakeIntRange(3, 5)
		reg.Size = tools.MakeIntRange(30, 50)
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{2, forest_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{3, mountain_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{1, river_generator.Create()})

		reg.ForcedGenerators = append(reg.ForcedGenerators, resource_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, city_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, road_generator.Create())

		regions[reg.Name] = reg
	}

	{
		var reg regionDefinition
		reg.Name = "Lakeland"
		reg.Usable = tools.MakeIntRange(3, 5)
		reg.Size = tools.MakeIntRange(30, 50)
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{2, forest_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{1, mountain_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{1, sea_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{3, river_generator.Create()})

		reg.ForcedGenerators = append(reg.ForcedGenerators, resource_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, city_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, road_generator.Create())

		regions[reg.Name] = reg
	}
	{
		var reg regionDefinition
		reg.Name = "Scorchinglands"
		reg.Usable = tools.MakeIntRange(1, 2)
		reg.Size = tools.MakeIntRange(30, 50)
		reg.Base = nodetype.Desert
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{1, forest_generator.Create()})
		reg.AvailableGenerators = append(reg.AvailableGenerators, generatorInclusion{3, mountain_generator.Create()})

		reg.ForcedGenerators = append(reg.ForcedGenerators, resource_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, city_generator.Create())
		reg.ForcedGenerators = append(reg.ForcedGenerators, road_generator.Create())

		regions[reg.Name] = reg
	}
}

// Generate a map generator based on region name
func Generate(name string) (*map_generator.MapGenerator, error) {
	reg, has := regions[name]
	if !has {
		return nil, errors.New("unknown region requested")
	}

	mgen := map_generator.New()
	mgen.Base = reg.Base

	nbIteration := reg.Usable.Roll()
	var gens []map_generator.MapSubGenerator
	for _, v := range reg.AvailableGenerators {
		for i := 0; i < v.Frequency; i++ {
			gens = append(gens, v.Generator)
		}
	}

	randomizer := tools.MakeIntRange(0, len(gens)-1)

	for i := 0; i < nbIteration; i++ {
		gen := gens[randomizer.Roll()]
		mgen.Generators[gen.Level()] = append(mgen.Generators[gen.Level()], gen)
	}
	for _, v := range reg.ForcedGenerators {
		mgen.Generators[v.Level()] = append(mgen.Generators[v.Level()], v)
	}

	return mgen, nil
}
