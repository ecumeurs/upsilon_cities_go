package resource_generator

import (
	"fmt"
	"log"
	"testing"
	"upsilon_cities_go/lib/cities/city/resource"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/map/pattern"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
)

func buildResource(depth int, proximity int, rarity int, exclusive bool) {
	var r resource.Resource
	r.Type = fmt.Sprintf("Test D%dP%d", depth, proximity)
	r.Name = fmt.Sprintf("Test Resource D%dP%d", depth, proximity)
	var c resource.Constraint
	c.NodeType = nodetype.Forest
	c.Depth = depth
	c.Proximity = proximity
	r.Constraints = append(r.Constraints, c)
	r.Rarity = rarity
	r.Exclusive = exclusive
	r.ID = len(DB)
	maxDist = tools.Max(maxDist, proximity)
	DB[len(DB)] = r
}

func shortMap() (gd *grid.CompoundedGrid) {
	gd = new(grid.CompoundedGrid)
	gd.Base = grid.Create(5, nodetype.Plain)
	gd.Delta = grid.Create(5, nodetype.None)
	return
}

func TestGatherResourcesD1P0(t *testing.T) {
	DB = make(map[int]resource.Resource)
	buildResource(1, 0, 1, false)
	gd := shortMap()

	// ensure our target practice is set.
	nd := gd.Get(node.NP(2, 2))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	dp := make(map[int]int)

	log.Printf("2,2: %v, pattern count %d", gd.Get(node.NP(2, 2)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(2, 2), gd, &dp)

	if len(res) != 1 {
		t.Errorf("Should have had test resource available. (%d)", len(res))
	}
}

func TestGatherResourcesD1P1(t *testing.T) {
	DB = make(map[int]resource.Resource)
	buildResource(1, 1, 1, false)
	gd := shortMap()

	nd := gd.Get(node.NP(2, 3))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	dp := make(map[int]int)

	log.Printf("2,2: %v, pattern count %d", gd.Get(node.NP(2, 2)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(2, 2), gd, &dp)

	if len(res) != 1 {
		t.Errorf("Should have had test resource available. (%d)", len(res))
	}
}

func TestComputeDepthD2(t *testing.T) {
	gd := shortMap()

	nds := gd.SelectPattern(node.NP(2, 3), pattern.GenerateAdjascentPattern(1))
	for _, nd := range nds {
		nd.Type = nodetype.Forest
		gd.Set(nd)
	}

	out := pattern.GenerateAdjascentOutlinePattern(2)
	log.Printf("outline: %v", out)

	var dm string
	dm = "\n"
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			dm += fmt.Sprintf("%d ", computeDepth(gd.Get(node.NP(i, j)), gd))
		}
		dm += "\n"
	}
	log.Printf(dm)
	log.Printf(gd.Compact().String())

	d := computeDepth(gd.Get(node.NP(2, 3)), gd)
	if d != 2 {
		t.Errorf("Should have had a depth of 2 god %d", d)
	}
}

func TestGatherResourcesD2P1(t *testing.T) {
	DB = make(map[int]resource.Resource)
	buildResource(2, 1, 1, false)
	gd := shortMap()

	nd := gd.Get(node.NP(2, 3))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	nds := gd.SelectPattern(node.NP(2, 3), pattern.GenerateAdjascentPattern(1))
	for _, nd := range nds {
		nd.Type = nodetype.Forest
		gd.Set(nd)
	}

	dp := make(map[int]int)

	log.Printf("2,2: %v, pattern count %d", gd.Get(node.NP(2, 2)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(2, 2), gd, &dp)

	if len(res) != 1 {
		t.Errorf("Should have had test resource available. (%d)", len(res))
	}
}

func TestGatherResourcesMixed(t *testing.T) {
	DB = make(map[int]resource.Resource)
	buildResource(2, 1, 1, false)
	buildResource(1, 1, 1, false)
	gd := shortMap()

	nd := gd.Get(node.NP(2, 3))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	nds := gd.SelectPattern(node.NP(2, 3), pattern.GenerateAdjascentPattern(1))
	for _, nd := range nds {
		nd.Type = nodetype.Forest
		gd.Set(nd)
	}

	dp := make(map[int]int)

	log.Printf("2,2: %v, pattern count %d", gd.Get(node.NP(2, 2)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(2, 2), gd, &dp)

	if len(res) != 2 {
		t.Errorf("Should have had test resource available. (%d)", len(res))
	}
}

func TestGatherResourcesD0P1(t *testing.T) {
	//exclusion test
	DB = make(map[int]resource.Resource)
	buildResource(0, 1, 1, false)
	gd := shortMap()

	nd := gd.Get(node.NP(2, 3))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	nds := gd.SelectPattern(node.NP(2, 3), pattern.GenerateAdjascentPattern(1))
	for _, nd := range nds {
		nd.Type = nodetype.Forest
		gd.Set(nd)
	}

	dp := make(map[int]int)

	log.Printf("2,2: %v, pattern count %d", gd.Get(node.NP(2, 2)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(2, 2), gd, &dp)

	if len(res) != 0 {
		t.Errorf("Should not have had test resource available. (%d)", len(res))
	}
}

func TestGatherResourcesD0P1T2(t *testing.T) {
	//exclusion test, this one should succeed
	DB = make(map[int]resource.Resource)
	buildResource(0, 1, 1, false)
	gd := shortMap()

	nd := gd.Get(node.NP(2, 3))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	nds := gd.SelectPattern(node.NP(2, 3), pattern.GenerateAdjascentPattern(1))
	for _, nd := range nds {
		nd.Type = nodetype.Forest
		gd.Set(nd)
	}

	dp := make(map[int]int)

	log.Printf("1,1: %v, pattern count %d", gd.Get(node.NP(1, 1)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(1, 1), gd, &dp)

	if len(res) != 1 {
		t.Errorf("Should not have had test resource available. (%d)", len(res))
	}
}

func TestGatherResources2C(t *testing.T) {
	DB = make(map[int]resource.Resource)
	var r resource.Resource
	r.Type = fmt.Sprintf("Test C")
	r.Name = fmt.Sprintf("Test C")
	var c resource.Constraint
	c.NodeType = nodetype.Forest
	c.Depth = 2
	c.Proximity = 2
	r.Constraints = append(r.Constraints, c)
	var c2 resource.Constraint
	c2.NodeType = nodetype.Plain
	c2.Depth = 1
	c2.Proximity = 2
	r.Constraints = append(r.Constraints, c2)
	r.Rarity = 1
	r.Exclusive = true
	r.ID = len(DB)
	maxDist = 2
	DB[len(DB)] = r

	gd := shortMap()

	nd := gd.Get(node.NP(2, 3))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	nds := gd.SelectPattern(node.NP(2, 3), pattern.GenerateAdjascentPattern(1))
	for _, nd := range nds {
		nd.Type = nodetype.Forest
		gd.Set(nd)
	}

	dp := make(map[int]int)

	log.Printf("2,3: %v, pattern count %d", gd.Get(node.NP(2, 3)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(2, 3), gd, &dp)

	if len(res) != 1 {
		t.Errorf("Should not have had test resource available. (%d)", len(res))
	}
}

func TestGatherResources2CFail(t *testing.T) {
	DB = make(map[int]resource.Resource)
	var r resource.Resource
	r.Type = fmt.Sprintf("Test C")
	r.Name = fmt.Sprintf("Test C")
	var c resource.Constraint
	c.NodeType = nodetype.Forest
	c.Depth = 3
	c.Proximity = 2
	r.Constraints = append(r.Constraints, c)
	var c2 resource.Constraint
	c2.NodeType = nodetype.Plain
	c2.Depth = 1
	c2.Proximity = 2
	r.Constraints = append(r.Constraints, c2)
	r.Rarity = 1
	r.Exclusive = true
	r.ID = len(DB)
	maxDist = 2
	DB[len(DB)] = r

	gd := shortMap()

	nd := gd.Get(node.NP(2, 3))
	nd.Type = nodetype.Forest
	gd.Set(nd)

	nds := gd.SelectPattern(node.NP(2, 3), pattern.GenerateAdjascentPattern(1))
	for _, nd := range nds {
		nd.Type = nodetype.Forest
		gd.Set(nd)
	}

	dp := make(map[int]int)

	log.Printf("2,3: %v, pattern count %d", gd.Get(node.NP(2, 3)).Type.String(), len(gd.SelectPattern(node.NP(2, 2), pattern.GenerateAdjascentPattern(0))))
	res := GatherResourcesAvailable(node.NP(2, 3), gd, &dp)

	if len(res) != 0 {
		t.Errorf("Should not have had test resource available. (%d)", len(res))
	}
}
