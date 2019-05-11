package storage

import (
	"fmt"
	"log"
	"testing"
	"upsilon_cities_go/lib/cities/item"
)

var itemNameID int

func generateItem() (res item.Item) {

	itemNameID++
	res.Name = fmt.Sprintf("Some Item %d", itemNameID)
	res.Type = "Some Item type"
	res.Quality = 10
	res.Quantity = 5
	res.BasePrice = 10

	return
}

func (storage *Storage) state() {
	log.Printf("Store State \n%s", storage.Pretty())
}

func TestAddToStorage(t *testing.T) {
	store := New()
	itm := generateItem()

	store.Add(itm)

	nitem, found := store.Get(1)
	if !found {
		t.Errorf("Failed to add item in store")
		store.state()
		return
	}
	if !itm.Match(nitem) {
		t.Errorf("Found an item in place but isn't the expected one.")
		store.state()
		return
	}
}

func TestAddToStorageFillStack(t *testing.T) {
	store := New()
	itm := generateItem()

	store.Add(itm)
	store.Add(itm)

	nitem, found := store.Get(1)
	if !found {
		t.Errorf("Failed to add item in store")
		store.state()
		return
	}
	if !itm.Match(nitem) {
		t.Errorf("Found an item in place but isn't the expected one.")
		store.state()
		return
	}
	if nitem.Quantity != itm.Quantity*2 {
		t.Errorf("Item in store doesn't have expected quantity (got: %d, expected %d)", nitem.Quantity, itm.Quantity*2)
		store.state()
		return
	}
}

func TestAddToStorageFailByCapacity(t *testing.T) {
	store := New()
	itm := generateItem()
	itm.Quantity = 999

	err := store.Add(itm)
	if err == nil {
		t.Errorf("Expected to fail by capacity")
		store.state()
		return
	}
}

func TestRemoveFromStorageEmptiesStack(t *testing.T) {
	store := New()
	itm := generateItem()

	store.Add(itm)
	store.Add(itm)
	store.Remove(1, 5)

	nitem, found := store.Get(1)
	if !found {
		t.Errorf("Failed to add item in store")
		store.state()
		return
	}
	if !itm.Match(nitem) {
		t.Errorf("Found an item in place but isn't the expected one.")
		store.state()
		return
	}
	if nitem.Quantity != (itm.Quantity*2 - 5) {
		t.Errorf("Item in store doesn't have expected quantity (got: %d, expected %d)", nitem.Quantity, (itm.Quantity*2 - 5))
		store.state()
		return
	}
}

func TestRemoveFromStorageFailNotEnough(t *testing.T) {
	store := New()
	itm := generateItem()

	store.Add(itm)
	err := store.Remove(1, itm.Quantity*2)

	if err == nil {
		t.Errorf("Expected an error to occurs when removing more than expected items.")
		store.state()
		return
	}

}

func TestRemoveFromStorageDropItemWhenEmpty(t *testing.T) {
	store := New()
	itm := generateItem()

	store.Add(itm)
	store.Remove(1, itm.Quantity)

	// Item should have been awarded ID 1

	_, found := store.Get(1)
	if found {
		t.Errorf("Found item when it was expected to be missing.")
		store.state()
		return
	}
}

func TestFindFirstMatching(t *testing.T) {
	store := New()
	for i := 0; i < 10; i++ {
		itm := generateItem()
		store.Add(itm)
	}
	itm := generateItem()
	store.Add(itm)

	fitm := store.First(func(item item.Item) bool {
		return item.Type == itm.Type
	})

	if fitm == nil {
		t.Errorf("Expected an item to be found matching %s", itm.Pretty())
		store.state()
		return
	}

	if !itm.Match(*fitm) {
		t.Errorf("An item has been found, but doesn't match requirement. %s vs %s", itm.Pretty(), fitm.Pretty())
		store.state()
		return
	}

}
