package resource_generator

import (
	"upsilon_cities_go/lib/cities/city/resource"
	rg "upsilon_cities_go/lib/cities/city/resource_generator"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/map_generator/map_level"
	"upsilon_cities_go/lib/cities/tools"
)

//ResourceGenerator generate resource ahah
type ResourceGenerator struct {
}

//Create a new resource generator with randomized conf
func Create() (mg ResourceGenerator) {
	return
}

//Level of the sub generator see Generator Level
func (mg ResourceGenerator) Level() map_level.GeneratorLevel {
	return map_level.Resource
}

func expandResources(rs []resource.Resource) (res []resource.Resource) {
	for _, v := range rs {
		for i := 0; i < v.Rarity; i++ {
			res = append(res, v)
		}
	}
	return
}

//Generate Will apply generator to provided grid
func (mg ResourceGenerator) Generate(gd *grid.CompoundedGrid) error {

	// quite simple:
	// for each cell, capture check ressource available, roll 1 that will be "available" immediately, then place all other
	// in the potential stack of the node with a rarity uped by 4
	// None typed ressources are removed from potential, of course.
	// if none is rolled, then node will have nothing available as a starter.

	acc := gd.AccessibilityGrid()
	depths := make(map[int]int) // depth ... will be computed only once ;) that's a memoisation stuff

	for idx := range acc.AvailableCells {
		availableResources := rg.GatherResourcesAvailable(acc.AvailableCells[idx], gd, &depths)
		expandedResources := expandResources(availableResources)
		tidx := tools.RandInt(0, len(expandedResources)-1)
		rsce := expandedResources[tidx]
		nd := gd.Get(acc.AvailableCells[idx])
		if rsce.Type == "None" {
			// shame ;)

			nd.Potential = availableResources
		} else {
			nd.Activated = append(nd.Activated, rsce)
			// remove used ressource from available ...
			navailable := make([]resource.Resource, 0, len(availableResources)-1)
			for ridx := range availableResources {
				if availableResources[ridx].Type != rsce.Type {
					rsc := availableResources[ridx]
					rsc.Clean()
					navailable = append(navailable, rsc)

				}
			}
			nd.Potential = navailable
		}

	}

	return nil
}

//Name of the generator
func (mg ResourceGenerator) Name() string {
	return "ResourceGenerator"
}
