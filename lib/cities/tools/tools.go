package tools

import "math/rand"

//IntRange range of int ;)
type IntRange struct {
	Min int
	Max int
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

//Roll Random in intrange.
func (ir IntRange) Roll() int {
	return rand.Intn(ir.Max-ir.Min) + ir.Min
}
