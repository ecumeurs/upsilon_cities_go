package producer

import (
	"fmt"
	"strings"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
)

type requirement struct {
	RessourceType string
	Quality       tools.IntRange
	Quantity      int
}

//Producer tell what it produce, within which criteria
type Producer struct {
	Product      item.Item
	Quality      tools.IntRange
	Quantity     tools.IntRange
	BasePrice    tools.IntRange
	Requirements []requirement
	Delay        int // in cycles
	Level        int // mostly informative, as levels will be applied directly to ranges, requirements and delay
}

//Produce create a new item based on template
func (prod *Producer) Produce() (res item.Item) {
	res = prod.Product
	res.Quality = prod.Quality.Roll()
	res.Quantity = prod.Quantity.Roll()
	res.BasePrice = prod.BasePrice.Roll()
	return
}

func (rq requirement) String() string {
	return fmt.Sprintf("%d x %s Q[%d-%d]", rq.Quantity, rq.RessourceType, rq.Quality.Min, rq.Quality.Max)
}

//CanProduce tell whether it's able to produce item, if it can produce it relayabely or if not, tell why.
func CanProduce(store *storage.Storage, prod *Producer, ressourcesGenerators []*Producer) (producable bool, nb int, recurrent bool, err error) {
	// producable immediately ?

	missing := make([]string, 0)
	found := make(map[string]int)
	for _, v := range prod.Requirements {
		found[v.RessourceType] = 0

		for _, foundling := range store.All(func(it item.Item) bool { return it.Type == v.RessourceType && tools.InEqRange(it.Quality, v.Quality) }) {
			found[v.RessourceType] += foundling.Quantity
		}

		if found[v.RessourceType] < v.Quantity {
			missing = append(missing, v.String())
		}
	}

	if len(missing) > 0 {
		return false, 0, false, fmt.Errorf("not enough ressources: %s", strings.Join(missing, ", "))
	}

	producable = true
	// we can at least produce one this.

	available := make(map[string]int)
	// check if we can produce more than one

	for _, v := range prod.Requirements {
		available[v.RessourceType] = found[v.RessourceType] / v.Quantity
	}

	nb = 0

	for _, v := range available {
		nb = tools.Max(nb, v)
	}

	// check if we can produce indefinitely ( due to ressource generator => production exceed requirements )

	found = make(map[string]int)
	available = make(map[string]int)
	for _, v := range prod.Requirements {
		for _, gen := range ressourcesGenerators {
			if gen.Product.Type == v.RessourceType {
				if gen.Quality.Min >= v.Quality.Min {
					found[v.RessourceType] += gen.Quantity.Min
				}
			}
		}
	}

	for _, v := range prod.Requirements {
		if _, has := found[v.RessourceType]; has {
			available[v.RessourceType] = found[v.RessourceType] / v.Quantity
		} else {
			return producable, nb, false, nil
		}
	}

	return producable, nb, true, nil
}
