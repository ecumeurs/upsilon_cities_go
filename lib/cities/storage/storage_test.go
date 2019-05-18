package storage

import (
	"fmt"
	"log"
	"testing"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/tools"
)

var itemNameID int

func generateItem() (res item.Item) {

	itemNameID++
	res.Name = fmt.Sprintf("Some Item %d", itemNameID)
	res.Type = []string{"Some Item type"}
	res.Quality = 10
	res.Quantity = 5
	res.BasePrice = 10

	return
}

func (storage *Storage) state() {
	log.Printf("Store State \n%s", storage.Pretty())
}

func TestForceFindFirstMatch(t *testing.T) {
	// forcefully fill store ...

	store := New()
	itm := generateItem()
	itm.ID = store.CurrentMaxID
	store.CurrentMaxID++
	store.Content[itm.ID] = itm

	storedit, err := store.First(ByMatch(itm))

	if err != nil {
		t.Errorf("unable to find in store itm %v", itm)
		store.state()
		return
	}

	if !itm.Match(storedit) {
		t.Errorf("found item doesn't match %v vs %v", itm, storedit)
		store.state()
		return
	}
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
	store.SetSize(100)
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
	store.SetSize(100)
	for i := 0; i < 10; i++ {
		itm := generateItem()
		itm.Type = []string{fmt.Sprintf("%s %d", itm.Type[0], i)}
		store.Add(itm)
	}
	itm := generateItem()
	store.Add(itm)

	fitm, err := store.First(ByType(itm.Type[0]))

	if err != nil {
		t.Errorf("Expected an item to be found matching %s", itm.Pretty())
		store.state()
		return
	}

	if !itm.Match(fitm) {
		t.Errorf("an item has been found, but doesn't match requirement. %s vs %s", itm.Pretty(), fitm.Pretty())
		store.state()
		return
	}
}
func TestFindAllMatching(t *testing.T) {
	store := New()
	store.SetSize(100)
	for i := 0; i < 10; i++ {
		itm := generateItem()
		itm.Type = []string{fmt.Sprintf("%s %d", itm.Type[0], i)}
		store.Add(itm)
	}

	// ensure 3 items of same type, but in different stacks ( by quality )
	itm := generateItem()
	store.Add(itm)
	itm.Quality += 3
	store.Add(itm)
	itm.Quality += 3
	store.Add(itm)

	fitm := store.All(ByType(itm.Type[0]))

	if len(fitm) == 0 {
		t.Errorf("Expected an item to be found matching %s", itm.Pretty())
		store.state()
		return
	}

	if len(fitm) != 3 {
		t.Errorf("an item has been found, but found not enough of them.")
		fmt.Printf("Expected %v got %v", itm, fitm)
		store.state()
		return
	}

	for _, v := range fitm {
		if !tools.StringListMatch(itm.Type, v.Type) {
			t.Errorf("an item has been found, but doesn't match requirement. %s vs %s", itm.Pretty(), v.Pretty())
			store.state()
			return
		}
	}
}

func TestReserveStorageSpace(t *testing.T) {
	store := New()

	_, err := store.Reserve(5)

	if err != nil {
		t.Errorf("didn't expect there to be an error while reserving space")
		return
	}

	if store.Count() != 5 {
		t.Errorf("expected Store to have some space used (got: %d ), but hasn't got.", store.Count())
		return
	}
}

func TestFailReserveStorageSpace(t *testing.T) {
	store := New()

	_, err := store.Reserve(15)

	if err == nil {
		t.Errorf("Expected to fail to claim reserved space but hasn't")
		store.state()
		return
	}

	if store.Count() != 0 {
		t.Errorf("Expected Store to be empty (as reserve failed) but isn't")
		return
	}
}

func TestClaimStorageSpace(t *testing.T) {
	store := New()

	id, _ := store.Reserve(5)

	itm := generateItem()

	itm.Quantity = 5

	var items []item.Item
	items = append(items, itm)

	err := store.Claim(id, items)

	if err != nil {
		t.Errorf("Failed to claim reserved space %d", id)
		store.state()
		return
	}

	if store.Count() != 5 {
		t.Errorf("Expected Store to have only our once reserved space used (expected: %d, got %d)", itm.Quantity, store.Count())
		return
	}
}

func TestGiveBackStoreSpace(t *testing.T) {
	store := New()

	id, _ := store.Reserve(5)
	err := store.GiveBack(id)

	if err != nil {
		t.Errorf("Failed to give back reserved space %d", id)
		store.state()
		return
	}

	if store.Count() != 0 {
		t.Errorf("Expected Store to be empty but isn't (expected: 0, got %d)", store.Count())
		return
	}

}
