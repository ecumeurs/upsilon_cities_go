package grid

import (
	"testing"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/nodetype"
)

func TestSetValueOfGridNode(t *testing.T) {
	gd := Create(40, nodetype.Plain)

	gd.GetP(10, 10).Type = nodetype.Mountain

	if gd.GetP(10, 10).Type != nodetype.Mountain {
		t.Error("P 10,10 should have been filled")
	}
}

func TestCompoundedGridIsFilled(t *testing.T) {
	gd := new(CompoundedGrid)
	gd.Base = Create(40, nodetype.Plain)
	gd.Delta = Create(40, nodetype.None)

	if gd.IsFilled(node.NP(10, 10)) {
		t.Error("P 10,10 shouldn't be filled has of yet")
	}

	gd.SetP(10, 10, nodetype.Mountain)

	if gd.Delta.GetP(10, 10).Type != nodetype.Mountain {
		t.Error("P 10,10 Delta should have been set to Mountain")
	}

	res := gd.Compact()

	if res.GetP(10, 10).Type != nodetype.Mountain {
		t.Error("P 10,10 Compacted should have been set to Mountain")
	}

	gd.Base = res

	if !gd.IsFilled(node.NP(10, 10)) {
		t.Error("P 10,10 should have been filled")
	}

}

func TestAccessibilityMoutain(t *testing.T) {
	base := Create(20, nodetype.Plain)

	// case 1
	// P P P
	// P M P
	// P P P
	base.GetP(10, 10).Type = nodetype.Mountain

	ag := base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( isolated mountain )")
		return
	}

	// case 2
	// M M M
	// M M M
	// M M M
	base.GetP(9, 9).Type = nodetype.Mountain
	base.GetP(10, 9).Type = nodetype.Mountain
	base.GetP(11, 9).Type = nodetype.Mountain
	base.GetP(9, 10).Type = nodetype.Mountain
	base.GetP(10, 10).Type = nodetype.Mountain
	base.GetP(11, 10).Type = nodetype.Mountain
	base.GetP(9, 11).Type = nodetype.Mountain
	base.GetP(10, 11).Type = nodetype.Mountain
	base.GetP(11, 11).Type = nodetype.Mountain

	ag = base.DefaultAccessibilityGrid()
	if ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should not be accessible ( surrounded mountain )")
		t.Errorf(ag.String())
		return
	}

	// case 3
	// M M M
	// P M M
	// M M M
	base.GetP(9, 10).Type = nodetype.Plain

	ag = base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( worst acceptable mountain )")
		return
	}
}

func TestAccessibilityForest(t *testing.T) {
	base := Create(20, nodetype.Plain)

	// case 1 forest alone
	// P P P
	// P F P
	// P P P
	base.GetP(10, 10).Type = nodetype.Forest

	ag := base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( isolated Forest )")
		return
	}

	// case 2 forest surrounded

	base.FillSquare(nodetype.Forest, 3, node.NP(10, 10))

	ag = base.DefaultAccessibilityGrid()
	if ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should not be accessible ( surrounded forest )")
		return
	}

	// case 3 worst case forest
	// F F F
	// P F F
	// F F F
	base.FillSquare(nodetype.Forest, 3, node.NP(10, 10))
	base.GetP(9, 10).Type = nodetype.Plain
	base.GetP(8, 10).Type = nodetype.Plain
	base.GetP(7, 10).Type = nodetype.Plain

	ag = base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( worst acceptable forest )")
		return
	}
}

func TestAccessibilityInaccessible(t *testing.T) {
	base := Create(20, nodetype.Plain)

	base.FillSquare(nodetype.Mountain, 6, node.NP(10, 10))

	ag := base.DefaultAccessibilityGrid()

	if ag.IsUsable() {
		t.Errorf("Map shouldn't be usable")
		return
	}
}
func TestAccessibilityAccessible(t *testing.T) {
	base := Create(20, nodetype.Plain)

	base.FillSquare(nodetype.Mountain, 5, node.NP(10, 10))

	ag := base.DefaultAccessibilityGrid()

	if !ag.IsUsable() {
		t.Errorf("Map should be usable")
		return
	}
}
