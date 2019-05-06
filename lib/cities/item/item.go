package item

//Item is a beautifull item
type Item struct {
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
