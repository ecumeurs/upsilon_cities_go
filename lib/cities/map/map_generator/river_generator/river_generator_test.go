package river_generator

import (
	"log"
	_ "net/http/pprof"
	"testing"
	"upsilon_cities_go/lib/cities/map/grid"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/system"
)

func TestRiverGenerator(t *testing.T) {

	system.LoadConf()
	dbh := db.NewTest()
	db.FlushDatabase(dbh)
	rg := Create()
	rg.Directness = tools.MakeIntRange(0, 0) // super direct to begin with ;)
	rg.Length = tools.MakeIntRange(10, 10)   // 10 cells in length and that's it !
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)

	// force a mountain
	gd.Base.Get(node.NP(5, 5)).Landscape = nodetype.Mountain
	// force a sea
	gd.Base.Get(node.NP(5, 15)).Ground = nodetype.Sea

	gd.Delta = grid.Create(20, nodetype.NoGround)

	rg.Generate(gd, dbh)
}

func TestRiverGeneratorDirectness(t *testing.T) {

	system.LoadConf()
	dbh := db.NewTest()
	db.FlushDatabase(dbh)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	rg := Create()
	rg.Directness = tools.MakeIntRange(3, 3) // super direct to begin with ;)
	rg.Length = tools.MakeIntRange(10, 10)   // 10 cells in length and that's it !
	gd := new(grid.CompoundedGrid)
	gd.Base = grid.Create(20, nodetype.Plain)

	// force a mountain
	gd.Base.Get(node.NP(5, 5)).Landscape = nodetype.Mountain
	// force a sea
	gd.Base.Get(node.NP(5, 15)).Ground = nodetype.Sea

	gd.Delta = grid.Create(20, nodetype.NoGround)

	rg.Generate(gd, dbh)

	// now there should be a river from 5,5 to 5,15

	t.Errorf(gd.Base.String())
	t.Errorf(gd.Delta.String())

}
