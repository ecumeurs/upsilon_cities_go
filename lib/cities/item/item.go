package item

import (
	"fmt"
	"log"
)

//Item is a beautifull item
type Item struct {
	ID        int64
	Name      string
	Type      string
	Quality   int
	Quantity  int
	BasePrice int // at quality 100
}

//Price compute a price, should provide an segmented valuation stuff ;)
func (it Item) Price() int {
	return it.BasePrice * (it.Quality / 100)
}

//Match tell whether two item are same(almost)
func (lhs Item) Match(rhs Item) bool {
	return lhs.Type == rhs.Type && lhs.Quality == rhs.Quality
}

//Pretty string
func (v Item) Pretty() string {
	return fmt.Sprintf("%d: %s (%s) Q[%d] x %d", v.ID, v.Name, v.Type, v.Quality, v.Quantity)
}

func (it Item) State() {
	log.Printf("Item: %s", it.Pretty())
}
