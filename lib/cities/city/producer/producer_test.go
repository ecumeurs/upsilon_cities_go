package producer

import (
	"testing"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
)

func generateItem(itemtype string) (res item.Item) {

	res.Name = itemtype
	res.Type = []string{itemtype}
	res.Quality = 10
	res.Quantity = 5
	res.BasePrice = 10

	return
}

func generateRessourceProducer() (prod Producer) {

	prod.ID = 0 // unset right now, will be the job of City to assign it an id.
	prod.ProductName = "Fruit"
	prod.ProductType = []string{"Fruit"}
	prod.Name = "Orchard"
	prod.Quality = tools.IntRange{Min: 10, Max: 20}
	prod.Quantity = tools.IntRange{Min: 10, Max: 20}
	prod.Delay = 20
	prod.BasePrice = 100
	prod.Level = 1
	prod.CurrentXP = 0
	prod.NextLevel = 100
	return prod

}

func generateFactoryProducer() (prod Producer) {

	myrange := tools.IntRange{Min: 10, Max: 20}
	prod.ID = 0 // unset right now, will be the job of City to assign it an id.
	prod.ProductName = "Pie"
	prod.ProductType = []string{"Food"}
	prod.Name = "Bakery"
	prod.Quality = myrange
	prod.Quantity = tools.IntRange{Min: 4, Max: 10}
	prod.Delay = 20
	prod.Requirements = append(prod.Requirements, requirement{Ressource: "Fruit", Type: true, Quality: myrange, Quantity: 7})
	prod.Requirements = append(prod.Requirements, requirement{Ressource: "Spice", Type: true, Quality: myrange, Quantity: 1})
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

	produce, storageFull, _ := CanProduceShort(store, &myproducer)

	if produce {
		t.Errorf("Can produce when Storage full")
		return
	}
	if !storageFull {
		t.Errorf("Can produce and storage full isn't set.")
		return
	}

	store = storage.New()
	produce, _, _ = CanProduceShort(store, &myproducer)

	if !produce {
		t.Errorf("Can't produce on empty Storage")
		return
	}
	if storageFull {
		t.Errorf("Storage full is set and shouldn't")
		return
	}

}

func TestFactoryProducerCanProduce(t *testing.T) {

	// need 7 fruits + 1 spice
	myproducer := generateFactoryProducer()
	store := storage.New()

	store.SetSize(40)

	wood := generateItem("Wood")
	wood.Quantity = 1

	store.Add(wood)

	fruit := generateItem("Fruit")
	fruit.Quality = 11
	fruit.Quantity = 3

	store.Add(fruit)

	fruit = generateItem("Fruit")
	fruit.Quality = 12
	fruit.Quantity = 4

	// added 7 fruits Q 10-20

	store.Add(fruit)

	produce, _, err := CanProduceShort(store, &myproducer)

	if produce {
		t.Errorf("Shouldn't be able to produce ... lacks a Spice ")
		return
	}

	spice := generateItem("Spice")
	spice.Quantity = 1

	store.Add(spice)

	produce, _, err = CanProduceShort(store, &myproducer)

	if !produce {
		t.Errorf("Should be able to produce, but can't %v", err)
		return
	}

}

func TestFactoryRunOnName(t *testing.T) {
	myproducer := generateFactoryProducer()
	myproducer.Requirements = append(make([]requirement, 0), requirement{Ressource: "Ashwood", Type: false, Quality: tools.IntRange{Min: 5, Max: 10}, Quantity: 1})
	store := storage.New()

	store.SetSize(40)

	wood := generateItem("Wood")
	wood.Quantity = 1
	store.Add(wood)

	produce, _, err := CanProduceShort(store, &myproducer)

	if produce {
		t.Errorf("Shouldn't be able to produce ... lacks a Ashwood: %v ", err)
		return
	}

	wood = generateItem("Wood")
	wood.Name = "Ashwood"
	wood.Quantity = 1
	store.Add(wood)

	produce, _, err = CanProduceShort(store, &myproducer)
	if !produce {
		t.Errorf("Should be able to produce: %v ", err)
		return
	}
}

func TestMultiTypeFactory(t *testing.T) {
	// need 7 fruits + 1 spice
	myproducer := generateFactoryProducer()
	store := storage.New()

	store.SetSize(40)

	wood := generateItem("Wood")
	wood.Quantity = 1

	store.Add(wood)

	fruit := generateItem("Fruit")
	fruit.Type = []string{"Apple", "Fruit"}
	fruit.Quality = 11
	fruit.Quantity = 3

	store.Add(fruit)

	fruit = generateItem("Fruit")
	fruit.Type = []string{"Fruit", "Tomato"}
	fruit.Quality = 12
	fruit.Quantity = 4

	// added 7 fruits Q 10-20

	store.Add(fruit)

	spice := generateItem("Spice")
	spice.Quantity = 1

	store.Add(spice)

	produce, _, err := CanProduceShort(store, &myproducer)

	if !produce {
		t.Errorf("Should be able to produce, but can't %v", err)
		return
	}
}
