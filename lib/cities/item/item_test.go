package item

import (
	"fmt"
	"log"
	"testing"
)

var itemNameID int

func generateItem() (res Item) {

	itemNameID++
	res.Name = fmt.Sprintf("Some Item %d", itemNameID)
	res.Type = []string{fmt.Sprintf("Some Item Type %d", itemNameID)}
	res.Quality = 10
	res.Quantity = 5
	res.BasePrice = 10

	return
}

func (it Item) state() {
	log.Printf("Item: %s", it.Pretty())
}

func TestMatch(t *testing.T) {
	itm1 := generateItem()
	itm2 := itm1

	if !itm1.Match(itm2) {
		t.Errorf("Expected itm1 to match itm2")
		itm1.state()
		itm2.state()
		return
	}
}

func TestMissMatch(t *testing.T) {
	itm1 := generateItem()
	itm2 := generateItem()

	if itm1.Match(itm2) {
		t.Errorf("Expected itm1 to missmatch itm2")
		itm1.state()
		itm2.state()
		return
	}
}

func TestMissMatchQlt(t *testing.T) {
	itm1 := generateItem()
	itm2 := itm1
	itm2.Quality = itm2.Quality + 10

	if itm1.Match(itm2) {
		t.Errorf("Expected itm1 to missmatch itm2")
		itm1.state()
		itm2.state()
		return
	}
}
