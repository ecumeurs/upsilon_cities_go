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

func TestAccessibilityMoutain(t *testing.T) {
	base := Create(20, node.Plain)

	// case 1
	// P P P
	// P M P
	// P P P
	base.GetP(10, 10).Type = node.Mountain

	ag := base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( isolated mountain )")
		return
	}

	// case 2
	// M M M
	// M M M
	// M M M
	base.GetP(9, 9).Type = node.Mountain
	base.GetP(10, 9).Type = node.Mountain
	base.GetP(11, 9).Type = node.Mountain
	base.GetP(9, 10).Type = node.Mountain
	base.GetP(10, 10).Type = node.Mountain
	base.GetP(11, 10).Type = node.Mountain
	base.GetP(9, 11).Type = node.Mountain
	base.GetP(10, 11).Type = node.Mountain
	base.GetP(11, 11).Type = node.Mountain

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
	base.GetP(9, 10).Type = node.Plain

	ag = base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( worst acceptable mountain )")
		return
	}
}

func TestAccessibilityForest(t *testing.T) {
	base := Create(20, node.Plain)

	// case 1 forest alone
	// P P P
	// P F P
	// P P P
	base.GetP(10, 10).Type = node.Forest

	ag := base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( isolated Forest )")
		return
	}

	// case 2 forest surrounded

	base.FillSquare(node.Forest, 3, node.NP(10, 10))

	ag = base.DefaultAccessibilityGrid()
	if ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should not be accessible ( surrounded forest )")
		return
	}

	// case 3 worst case forest
	// F F F
	// P F F
	// F F F
	base.FillSquare(node.Forest, 3, node.NP(10, 10))
	base.GetP(9, 10).Type = node.Plain
	base.GetP(8, 10).Type = node.Plain
	base.GetP(7, 10).Type = node.Plain

	ag = base.DefaultAccessibilityGrid()
	if !ag.IsAccessibleP(10, 10) {
		t.Errorf("P 10,10 should be accessible ( worst acceptable forest )")
		return
	}
}

func TestAccessibilityInaccessible(t *testing.T) {
	base := Create(20, node.Plain)

	base.FillSquare(node.Mountain, 6, node.NP(10, 10))

	ag := base.DefaultAccessibilityGrid()

	if ag.IsUsable() {
		t.Errorf("Map shouldn't be usable")
		return
	}
}
func TestAccessibilityAccessible(t *testing.T) {
	base := Create(20, node.Plain)

	base.FillSquare(node.Mountain, 5, node.NP(10, 10))

	ag := base.DefaultAccessibilityGrid()

	if !ag.IsUsable() {
		t.Errorf("Map should be usable")
		return
	}
}

func TestAccessibilityExclusionDoesMatter(t *testing.T) {
	base := Create(20, node.Plain)

	base.AddLine(node.Mountain, node.NP(9, 9), node.NP(9, 20), 1)
	base.AddLine(node.Mountain, node.NP(9, 9), node.NP(20, 9), 1)

	ag := base.DefaultAccessibilityGrid()

	if ag.IsUsable() {
		t.Errorf("Map should not be usable")
		return
	}
}

func TestAccessibilityCantReachLockedOutZone(t *testing.T) {
	base := Create(20, node.Plain)

	base.AddLine(node.Mountain, node.NP(10, 10), node.NP(10, 20), 1)
	base.AddLine(node.Mountain, node.NP(9, 9), node.NP(20, 9), 1)

	ag := base.DefaultAccessibilityGrid()

	if !ag.IsUsable() {
		t.Errorf("Map should be usable")
		return
	}

	if ag.IsAccessibleP(15, 15) {
		t.Errorf("P 15,15 should not be accessible ( in a locked zone )")
		t.Errorf(ag.String())
		t.Errorf(base.String())
		return
	}
}