package tools

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

//IntRange range of int ;)
type IntRange struct {
	Min int
	Max int
}

//StringListMatchOne equal at least one
func StringListMatchOne(lhs, rhs []string) bool {
	for _, v := range lhs {
		for _, w := range rhs {
			if w == v {
				return true
			}
		}
	}
	return false
}

//StringListMatchAll equal all
func StringListMatchAll(lhs, rhs []string) bool {
	for _, v := range lhs {
		found := false
		for _, w := range rhs {
			if w == v {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

//InStringList is in list
func InStringList(value string, rhs []string) bool {
	for _, w := range rhs {
		if w == value {
			return true
		}
	}
	return false
}

//HasOneIn tell whether it as at least one item(of target) in list(ref).
func HasOneIn(ref []string, target []string) bool {
	for _, v := range target {
		if InStringList(v, ref) {
			return true
		}
	}
	return false
}

//OneIn get one item(of target) that is also in list(ref).
func OneIn(ref []string, target []string) (string, bool) {
	for _, v := range target {
		if InStringList(v, ref) {
			return v, true
		}
	}
	return "", false
}

//InList is in list
func InList(value int, rhs []int) bool {
	for _, w := range rhs {
		if w == value {
			return true
		}
	}
	return false
}

//ListInStringMap is in map; if any true, one match is sufficient.
func ListInStringMap(value []string, rhs map[string]bool, any bool) bool {
	found := true
	for _, v := range value {
		_, found = rhs[v]
		if found && any {
			return true
		} else if !found {
			return false
		}
	}
	return true
}

//ListInStringList is in list
func ListInStringList(value []string, rhs []string) bool {
	for _, v := range value {
		found := false
		for _, w := range rhs {
			if v == w {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

//Min returns min of two int
func Min(lhs, rhs int) int {
	if lhs < rhs {
		return lhs
	}
	return rhs
}

//Max returns max of two int
func Max(lhs, rhs int) int {
	if lhs > rhs {
		return lhs
	}
	return rhs
}

//In value in range
func In(value, left, right int) bool {
	return value > left && value < right
}

//InRange value in range
func InRange(value int, rg IntRange) bool {
	return value > rg.Min && value < rg.Max
}

//InEq value in range inclusive
func InEq(value, left, right int) bool {
	return value >= left && value <= right
}

//InEqRange value in range
func InEqRange(value int, rg IntRange) bool {
	return value >= rg.Min && value <= rg.Max
}

//Abs absolute value in int.
func Abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

//Floor floors a float to int.
func Floor(x float32) int {
	return int(math.Floor(float64(x)))
}

//Roll Random in intrange.
func (ir IntRange) Roll() int {
	if ir.Max-ir.Min == 0 {
		return ir.Min
	}
	return rand.Intn(ir.Max-ir.Min) + ir.Min
}

//MakeIntRange build an intrange.
func MakeIntRange(min int, max int) IntRange {
	return IntRange{min, max}
}

//RandInt random int
func RandInt(begin int, end int) int {
	return rand.Intn(end-begin) + begin
}

//CycleLength duration of a cycle
var CycleLength time.Duration

//InitCycle initialize cycle lenght.
func InitCycle() {
	CycleLength, _ = time.ParseDuration("10s")
}

//CyclesBetween Count cycles since then.
func CyclesBetween(t time.Time, t2 time.Time) int {
	if t.After(t2) {
		return int((t.Sub(t2)) / CycleLength)
	}
	return int((t2.Sub(t)) / CycleLength)
}

//RoundTime rounds up time up to cycle.
func RoundTime(base time.Time) time.Time {
	return base.Round(CycleLength)
}

//RoundNow rounds up now.
func RoundNow() time.Time {
	return time.Now().UTC().Round(CycleLength)
}

//AboutNow Alter now.
func AboutNow(cycles int) time.Time {
	return AddCycles(RoundNow(), cycles)
}

//AddCycles tell what time it will be in cycles cycles.
func AddCycles(base time.Time, cycles int) time.Time {
	return base.Round(CycleLength).Add(time.Duration(cycles) * CycleLength)
}

//MinTime return lesser time of the two
func MinTime(lhs time.Time, rhs time.Time) time.Time {
	if lhs.Before(rhs) {
		return lhs
	}
	return rhs
}

//MaxTime return lesser time of the two
func MaxTime(lhs time.Time, rhs time.Time) time.Time {
	if lhs.After(rhs) {
		return lhs
	}
	return rhs
}

//SortInt64 sort int64 array list
func SortInt64(arr []int64) {

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
}
