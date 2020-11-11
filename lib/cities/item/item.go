package item

import (
	"fmt"
	"log"
	"strings"
	"upsilon_cities_go/lib/cities/tools"
)

//Item is a beautifull item
type Item struct {
	ID        int64
	Name      string
	Type      []string
	Quality   int
	Quantity  int
	BasePrice int // at quality 100
}

//Price compute a price, should provide an segmented valuation stuff ;)
func (it Item) Price() int {
	return it.BasePrice * (it.Quality / 100)
}

//Match tell whether two item are same(almost)
func (it Item) Match(rhs Item) bool {
	return tools.StringListMatchAll(it.Type, rhs.Type) && it.Name == rhs.Name && it.Quality == rhs.Quality
}

//Pretty string
func (it Item) Pretty() string {
	return fmt.Sprintf("%d: %s (%s) Q[%d] x %d", it.ID, it.Name, it.Type, it.Quality, it.Quantity)
}

//ShortPretty string
func (it Item) ShortPretty() string {
	return fmt.Sprintf("%s (%s) Q[%d] x %d", it.Name, strings.Join(it.Type, ","), it.Quality, it.Quantity)
}

//PrettyTypes string
func (it Item) PrettyTypes() string {
	return strings.Join(it.Type, ",")
}

//State print item
func (it Item) State() {
	log.Printf("Item: %s", it.Pretty())
}
