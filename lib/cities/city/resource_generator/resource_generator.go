package resource_generator

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"upsilon_cities_go/lib/cities/city/resource"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/misc/config/system"
)

//DB Store in memory all ressource definitions.
var DB map[int]resource.Resource
var maxDist int

//Load resources from conf files.
func Load() {

	DB = make(map[int]resource.Resource)
	baseID := 0
	maxDist = 3

	filepath.Walk(system.MakePath(system.Get("data_resources", "data/resources")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Resource: prevent panic by handling failure accessing a path %q: %v\n", system.Get("data_producers", "data/producers"), err)
			return err
		}
		if strings.HasSuffix(info.Name(), ".json") {
			f, ferr := os.Open(path)
			if ferr != nil {
				log.Fatalln("Resource: No Resource: data file present")
			}

			producerJSON, ferr := ioutil.ReadAll(f)
			if ferr != nil {
				log.Fatalln("Resource: Data file found but unable to read it all.")
			}

			f.Close()

			prods := make([]resource.Resource, 0)
			json.Unmarshal(producerJSON, &prods)

			for _, p := range prods {
				baseID++
				p.ID = baseID
				for _, c := range p.Constraints {
					if c.Proximity > maxDist {
						maxDist = c.Proximity
					}
				}
				DB[baseID] = p
				log.Printf("Resource: Loaded Resource %v ", p.String())
			}
		}

		return nil
	})

}

func computeDepth(n node.Node, gd *grid.CompoundedGrid) (depth int) {
	depth = 1
	for true {
		test := gd.SelectPattern(n.Location, pattern.GenerateAdjascentOutlinePattern(depth+1))
		for _, v := range test {
			if v.Type != n.Type {
				return
			}
		}
		depth++
	}
	return
}

func matchConstraint(targets []node.Node, c resource.Constraint, depths *map[int]int, mapSize int) bool {
	if c.Depth == 0 {
		// this is an exclusion constraint, means, it shouldn't have any items with specified type.
		for _, t := range targets {
			if t.Type == c.NodeType {
				return false
			}
		}
	}

	for _, t := range targets {
		if v := (*depths)[t.Location.ToInt(mapSize)]; v >= c.Depth {
			return true
		}
	}

	return false
}

//GatherResourcesAvailable will gather all available resources based resource.DB
func GatherResourcesAvailable(loc node.Point, gd *grid.CompoundedGrid, depths *map[int]int) (res []resource.Resource) {
	availableDists := make(map[int][]node.Node)

	for i := 0; i < maxDist; i++ {
		availableDists[i] = gd.SelectPattern(loc, pattern.GenerateAdjascentPattern(i))
		for _, v := range availableDists[i] {
			if _, found := (*depths)[v.Location.ToInt(gd.Base.Size)]; !found {
				(*depths)[v.Location.ToInt(gd.Base.Size)] = computeDepth(v, gd)
			}
		}
	}

	tmpRes := make(map[string][]resource.Resource)

	for _, v := range DB {

		allOk := true
		for _, c := range v.Constraints {
			if !matchConstraint(availableDists[c.Proximity], c, depths, gd.Base.Size) {
				allOk = false
				break
			}
		}

		if allOk {
			tmpRes[v.Type] = append(tmpRes[v.Type], v)
		}
	}

	for _, v := range tmpRes {
		resResource := v[0]
		resResource.Rarity = 0
		for _, r := range v {
			if r.Exclusive {
				resResource = r
				break
			}

			resResource.Rarity += r.Rarity
		}
		res = append(res, resResource)
	}

	return
}
