package producer

import (
	"testing"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
)

func generateItem(itemtype string) (res item.Item) {

	res.Name = itemtype
	res.Type = itemtype
	res.Quality = 10
	res.Quantity = 5
	res.BasePrice = 10

	return
}

func generateRessourceProducer() (prod Producer) {

	prod.ID = 0 // unset right now, will be the job of City to assign it an id.
	prod.ProductName = "Fruit"
	prod.ProductType = "Fruit"
	prod.Name = "Orchard"
	prod.Quality = tools.IntRange{10, 20}
	prod.Quantity = tools.IntRange{10, 20}
	prod.Delay = 20
	prod.BasePrice = 100
	prod.Level = 1
	prod.CurrentXP = 0
	prod.NextLevel = 100
	return prod

}

func generateFactoryProducer() (prod Producer) {

	myrange := tools.IntRange{10, 20}
	prod.ID = 0 // unset right now, will be the job of City to assign it an id.
	prod.ProductName = "Pie"
	prod.ProductType = "Food"
	prod.Name = "Bakery"
	prod.Quality = myrange
	prod.Quantity = tools.IntRange{4, 10}
	prod.Delay = 20
	prod.Requirements = append(prod.Requirements, requirement{RessourceType: "Fruit", Quality: myrange, Quantity: 3})
	prod.Requirements = append(prod.Requirements, requirement{RessourceType: "Spice", Quality: myrange, Quantity: 1})
	prod.BasePrice = 100
	prod.Level = 1
	prod.CurrentXP = 0
	prod.NextLevel = 100
	return prod

}

func TestRessourceProducerCanProduce(t *testing.T) {

	myproducer := generateRessourceProducer()
	store := storage.New()
	itm := generateItem("someitem")
	itm.Quantity = 10

	store.Add(itm)

	produce := CanProduceShort(store, &myproducer)

	if produce {
		t.Errorf("Can produce when Storage full")
		return
	}

	store = storage.New()
	produce = CanProduceShort(store, &myproducer)

	if !produce {
		t.Errorf("Can't produce on empty Storage")
		return
	}

}

func TestFactoryProducerCanProduce(t *testing.T) {

	myproducer := generateFactoryProducer()
	store := storage.New()

	fruit := generateItem("Fruit")
	fruit.Quantity = 8

	store.Add(fruit)

	spice := generateItem("Spice")
	spice.Quantity = 2

	store.Add(spice)

	produce := CanProduceShort(store, &myproducer)

	if !produce {
		t.Errorf("Enough item for requierement : can't produce")
		return
	}

	store = storage.New()
	fruit = generateItem("Fruit")
	fruit.Quantity = 2

	store.Add(fruit)

	spice = generateItem("Spice")
	spice.Quantity = 8

	produce = CanProduceShort(store, &myproducer)

	if produce {
		t.Errorf("Not enough item for requierement : produce")
		return
	}

}
