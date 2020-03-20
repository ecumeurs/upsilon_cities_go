package resource_generator

import (
	"log"
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

	depths := make(map[int]int) // depth ... will be computed only once ;) that's a memoisation stuff

	for idx := range gd.Base.Nodes {
		availableResources := rg.GatherResourcesAvailable(gd.Base.Nodes[idx].Location, gd, &depths)
		expandedResources := expandResources(availableResources)
		if len(expandedResources) == 0 {
			// nothing to provide here, must be some deep see shit.

			continue
		}
		tidx := tools.RandInt(0, len(expandedResources)-1)
		rsce := expandedResources[tidx]
		nd := gd.Get(gd.Base.Nodes[idx].Location)

		log.Printf("RG: Node %v: available resources: %d", nd.Location.String(), len(availableResources))
		log.Printf("RG: activated: %s", rsce.Type)

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

		gd.Update(nd)

	}

	return nil
}

//Name of the generator
func (mg ResourceGenerator) Name() string {
	return "ResourceGenerator"
}
