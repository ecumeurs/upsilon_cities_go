package grid

import (
	"testing"
	"upsilon_cities_go/lib/cities/node"
)

func TestSetValueOfGridNode(t *testing.T) {
	gd := Create(40, node.Plain)

	gd.GetP(10, 10).Type = node.Mountain

	if gd.GetP(10, 10).Type != node.Mountain {
		t.Error("P 10,10 should have been filled")
	}
}

func TestCompoundedGridIsFilled(t *testing.T) {
	gd := new(CompoundedGrid)
	gd.Base = Create(40, node.Plain)
	gd.Delta = Create(40, node.None)

	if gd.IsFilled(node.NP(10, 10)) {
		t.Error("P 10,10 shouldn't be filled has of yet")
	}

	gd.SetP(10, 10, node.Mountain)

	if gd.Delta.GetP(10, 10).Type != node.Mountain {
		t.Error("P 10,10 Delta should have been set to Mountain")
	}

	res := gd.Compact()

	if res.GetP(10, 10).Type != node.Mountain {
		t.Error("P 10,10 Compacted should have been set to Mountain")
	}

	gd.Base = res

	if !gd.IsFilled(node.NP(10, 10)) {
		t.Error("P 10,10 should have been filled")
	}

}
