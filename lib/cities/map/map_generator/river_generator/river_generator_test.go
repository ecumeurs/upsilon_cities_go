package river_generator

import (
	"log"
	_ "net/http/pprof"
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
)

func TestRiverGenerator(t *testing.T) {

	rg := Create()
	rg.Directness = tools.MakeIntRange(0, 0) // super direct to begin with ;)
	rg.Length = tools.MakeIntRange(10, 10)   // 10 cells in length and that's it !
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)

	// force a mountain
	gd.Base.Get(node.NP(5, 5)).Type = nodetype.Mountain
	// force a sea
	gd.Base.Get(node.NP(5, 15)).Type = nodetype.Sea

	gd.Delta = grid.Create(20, nodetype.None)

	rg.Generate(gd)
}

func TestRiverGeneratorDirectness(t *testing.T) {

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	rg := Create()
	rg.Directness = tools.MakeIntRange(3, 3) // super direct to begin with ;)
	rg.Length = tools.MakeIntRange(10, 10)   // 10 cells in length and that's it !
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)

	// force a mountain
	gd.Base.Get(node.NP(5, 5)).Type = nodetype.Mountain
	// force a sea
	gd.Base.Get(node.NP(5, 15)).Type = nodetype.Sea

	gd.Delta = grid.Create(20, nodetype.None)

	rg.Generate(gd)

	// now there should be a river from 5,5 to 5,15

	t.Errorf(gd.Base.String())
	t.Errorf(gd.Delta.String())

}
