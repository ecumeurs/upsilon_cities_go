package grid

import (
	"log"
	"strconv"
	"upsilon_cities_go/lib/cities/node"
	"upsilon_cities_go/lib/cities/tools"
)

//AccessibilityGridStruct describe what's accessible
type AccessibilityGridStruct struct {
	Grid
	FillRate       float64
	AvailableCells []node.Point
	NbAvailable    int
}

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

//DefaultAccessibilityGrid generate a grid that store only accessibility informations
//Will accept any grid with at least a cluster bigger than 0.4
func (gd Grid) DefaultAccessibilityGrid() (res AccessibilityGridStruct) {
	return gd.AccessibilityGrid(0.7)
}

//AccessibilityGrid generate a grid that store only accessibility informations
//fillRatio decides when an isolated cluster of points is to be dropped altogether
func (gd Grid) AccessibilityGrid(fillRatio float64) (res AccessibilityGridStruct) {
	// forest away from plain for more than 3 tiles are deemed inaccessible
	// desert away from plain for more than 3 tiles are deemed inaccessible
	// mountains away from plain for more than 1 tile are deemed inaccessible
	// sea away from plain for more than 1 tile are deemed inaccessible ( for now )
	// inaccessible means that no road can be established there.
	// it also means that no cities may be put there.

	// init accessibility grid.
	res.Size = gd.Size
	for y := 0; y < gd.Size; y++ {
		for x := 0; x < gd.Size; x++ {
			n := node.New(x, y)
			n.ID = y*gd.Size + x
			n.Type = node.Accessible
			res.Nodes = append(res.Nodes, n)
		}
	}

	depth3 := make([]int, gd.Size*gd.Size)
	depth1 := make([]int, gd.Size*gd.Size)

	for y := 0; y < gd.Size; y++ {
		for x := 0; x < gd.Size; x++ {
			switch typ := gd.GetP(x, y).Type; typ {
			case node.Forest:
				fill(1, 3, x, y, &depth3, gd.Size)
			case node.Desert:
				fill(1, 3, x, y, &depth3, gd.Size)
			case node.Sea:
				fill(1, 1, x, y, &depth1, gd.Size)
			case node.Mountain:
				fill(1, 1, x, y, &depth1, gd.Size)
			default:
			}
		}
	}

	res.NbAvailable = 0

	for y := 0; y < gd.Size; y++ {
		for x := 0; x < gd.Size; x++ {
			if depth3[x+y*gd.Size] > 46 || depth1[x+y*gd.Size] > 8 {
				res.Nodes[x+y*gd.Size].Type = node.Inaccessible
			} else {
				res.AvailableCells = append(res.AvailableCells, node.NP(x, y))
				res.NbAvailable++
			}
		}
	}
	log.Printf("Accessibility: fillrate check: %f > %f", float64(res.NbAvailable)/float64(gd.Size*gd.Size), fillRatio)

	if float64(res.NbAvailable)/float64(gd.Size*gd.Size) < fillRatio {
		// Well not enough available cells anyway...
		res.AvailableCells = make([]node.Point, 0)
		res.NbAvailable = 0
		res.FillRate = 0
		return
	}

	// Check for isolated zones ...
	for res.NbAvailable != 0 {
		// take one
		nd := res.AvailableCells[0]
		total, rest, used := countConnected(nd, res.AvailableCells[1:])
		// is found cluster ... relevant ? => means is it the biggest one available right now ?
		// log.Printf("Accessibility: fillrate check: %f > %f total %d vs %d", (float64(total) / float64(gd.Size*gd.Size)), fillRatio, total, res.NbAvailable/2)
		if total >= res.NbAvailable/2 {
			// is it still significant enough
			if (float64(total) / float64(gd.Size*gd.Size)) > fillRatio {
				res.AvailableCells = used
				res.NbAvailable = total
				res.FillRate = float64(res.NbAvailable) / float64(gd.Size*gd.Size)
				return
			}
			// Well nothing available, has too many cluster < fillRatio
			// => That was the biggest one available on this cluster and it didn't reach exclude ratio ... thus ...
			res.FillRate = 0
			res.AvailableCells = make([]node.Point, 0)
			res.NbAvailable = 0
			return
		}
		// Cluster was too small
		res.AvailableCells = rest
		res.NbAvailable = len(rest)
		if (float64(res.NbAvailable) / float64(gd.Size*gd.Size)) <= fillRatio {
			// not enough tiles available to meet expectations
			res.FillRate = 0
			res.AvailableCells = make([]node.Point, 0)
			res.NbAvailable = 0
			return
		}

	}

	// Well nothing available, has too many cluster < fillRatio
	res.FillRate = 0
	return
}

func countConnected(cur node.Point, availables []node.Point) (total int, rest []node.Point, used []node.Point) {
	used = append(used, cur)
	current := make([]node.Point, 0)
	current = append(current, cur)
	return subCountConnected(current, availables, used, 1)
}

//subCountConnected tell how many points are connected to this one
func subCountConnected(current []node.Point, availables []node.Point, used []node.Point, currentCount int) (total int, rest []node.Point, _used []node.Point) {
	sub := make([]node.Point, 0)

	if len(current) == 0 {
		return currentCount, availables, used
	}

	cur := current[0]
	candidates := make([]node.Point, 0)
	for idx, v := range availables {
		// seek adj point.
		if cur.IsAdj(v) {
			candidates = append(candidates, v)
			used = append(used, v)
		} else {
			sub = append(sub, v)
		}

		lc := len(candidates)
		if lc == 3 || v.Y > cur.Y+1 {
			if lc > 0 {
				if idx+1 < len(availables) {
					return subCountConnected(append(current[1:], candidates...), append(sub, availables[idx+1:]...), used, currentCount+len(candidates))
				}
			}
			return subCountConnected(append(current[1:], candidates...), append(sub, availables[idx+1:]...), used, currentCount+len(candidates))
		}
	}

	return subCountConnected(append(current[1:], candidates...), sub, used, currentCount+len(candidates))
}

//IsUsable tell whether grid is usable (based on fillrate)
func (gd AccessibilityGridStruct) IsUsable() bool {
	return gd.FillRate != 0
}

//IsAccessibleP tell whether target is accessible or not.
func (gd AccessibilityGridStruct) IsAccessibleP(x, y int) bool {
	if gd.FillRate == 0 {
		return false
	}
	return gd.GetP(x, y).Type == node.Accessible
}

//IsAccessible tell whether target is accessible or not.
func (gd AccessibilityGridStruct) IsAccessible(loc node.Point) bool {
	if gd.FillRate == 0 {
		return false
	}
	return gd.Get(loc).Type == node.Accessible
}
