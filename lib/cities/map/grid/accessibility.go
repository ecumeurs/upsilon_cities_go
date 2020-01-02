package grid

import (
	"strconv"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

func fill(inc int, dist int, centerX int, centerY int, table *[]int, rowSize int) {
	for x := tools.Max(0, centerX-dist); x < tools.Min(rowSize, centerX+1+dist); x++ {
		for y := tools.Max(0, centerY-dist); y < tools.Min(rowSize, centerY+1+dist); y++ {
			(*table)[x+y*rowSize] += inc
		}
	}
}

func stringify(table []int, row int) string {
	var res string
	for i, n := range table {
		if i%row == 0 {
			res += "\n"
		}
		res += strconv.Itoa(n) + " "
	}
	return res
}

//AccessibilityGrid generate a grid that store only accessibility informations
func (gd Grid) AccessibilityGrid() (res Grid) {
	// forest away from plain for more than 3 tiles are deemed inaccessible
	// desert away from plain for more than 3 tiles are deemed inaccessible
	// mountains away from plain for more than 1 tile are deemed inaccessible
	// sea away from plain for more than 1 tile are deemed inaccessible ( for now )
	// inaccessible means that no road can be established there.
	// it also means that no cities may be put there.

	// init accessibility grid.
	res.Size = gd.Size
	for i := 0; i < gd.Size; i++ {
		for j := 0; j < gd.Size; j++ {
			n := node.New(j, i)
			n.ID = i*gd.Size + j
			n.Type = node.Accessible
			res.Nodes = append(res.Nodes, n)
		}
	}

	depth3 := make([]int, gd.Size*gd.Size)
	depth1 := make([]int, gd.Size*gd.Size)

	for i := 0; i < gd.Size; i++ {
		for j := 0; j < gd.Size; j++ {
			switch typ := gd.GetP(i, j).Type; typ {
			case node.Forest:
				fill(1, 3, i, j, &depth3, gd.Size)
			case node.Desert:
				fill(1, 3, i, j, &depth3, gd.Size)
			case node.Sea:
				fill(1, 1, i, j, &depth1, gd.Size)
			case node.Mountain:
				fill(1, 1, i, j, &depth1, gd.Size)
			default:
			}
		}
	}

	for i := 0; i < gd.Size; i++ {
		for j := 0; j < gd.Size; j++ {
			if depth3[i+j*gd.Size] > 46 || depth1[i+j*gd.Size] > 8 {
				res.Nodes[i+j*gd.Size].Type = node.Inaccessible
			}
		}
	}

	return
}

//IsAccessibleP tell whether target is accessible or not.
func (gd Grid) IsAccessibleP(x, y int) bool {
	return gd.GetP(x, y).Type == node.Accessible
}

//IsAccessible tell whether target is accessible or not.
func (gd Grid) IsAccessible(loc node.Point) bool {
	return gd.Get(loc).Type == node.Accessible
}
