package storage

import (
	"errors"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/tools"
)

//Storage contains every item of city and storage capacity. Note `json:"-"` means it won't be exported as json ...
type Storage struct {
	Capacity     int
	Content      map[int64]item.Item
	CurrentMaxID int64 `json:"-"`
}

//New storage !
func New() (res *Storage) {
	res = new(Storage)
	res.Content = make(map[int64]item.Item)
	res.CurrentMaxID = 1
	res.Capacity = 10
	return
}

//SetSize set storage capacity
func (storage *Storage) SetSize(nsize int) {
	if storage.Capacity < nsize {
		storage.Capacity = nsize
	}
}

//Count return the nb of item in Storage
func (storage *Storage) Count() int {
	var total int
	for _, item := range storage.Content {
		total += item.Quantity
	}

	return total
}

//Add item to storage
func (storage *Storage) Add(it item.Item) error {
	if storage.Spaceleft() < it.Quantity {
		return errors.New("unable to insert item, no space left")
	}
	done := false
	storedit := storage.First(func(lhs item.Item) bool { return lhs.Match(it) })
	if storedit != nil {
		done = true
		storedit.Quantity += it.Quantity
		storage.Content[storedit.ID] = *storedit
		return nil
	}

	if !done {
		storage.Content[storage.CurrentMaxID] = it
		storage.CurrentMaxID++
	}
	return nil
}

//Remove item from storage
func (storage *Storage) Remove(id int64, nb int) error {
	itm, found := storage.Content[id]
	if !found {
		return errors.New("unable to remove unknown item")
	}

	if itm.Quantity < nb {
		return errors.New("unable to remove requested amount of items")
	}

	itm.Quantity -= nb
	storage.Content[id] = itm
	return nil
}

//Isfull return a boolean if nb item reach capacity
func (storage *Storage) Isfull() bool {
	return storage.Count() == storage.Capacity
}

//Spaceleft return space left depending of capacity
func (storage *Storage) Spaceleft() int {
	return storage.Capacity - storage.Count()
}

//Has tell whether store has item requested in number.
func (storage *Storage) Has(itType string, itNb int) bool {
	for _, it := range storage.Content {
		if it.Type == itType {
			if it.Quantity >= itNb {
				return true
			}
			// don't return false quite yet as it could have multiple time the same item type.
		}
	}
	return false
}

//HasQQ tell whether store has item requested in Quantity and Quality.
func (storage *Storage) HasQQ(itType string, quantity int, quality tools.IntRange) bool {
	for _, it := range storage.Content {
		if it.Type == itType {
			if it.Quantity >= quantity {
				if tools.InEqRange(it.Quality, quality) {
					return true
				}
			}
			// don't return false quite yet as it could have multiple time the same item type.
		}
	}
	return false
}

//HasCustom whether has item type matching requirement.
func (storage *Storage) HasCustom(itType string, tester func(item.Item) bool) bool {
	for _, it := range storage.Content {
		if it.Type == itType {
			if tester(it) {
				return true
			}
		}
	}
	return false
}

//All gather all items matching requirements
func (storage *Storage) All(tester func(item.Item) bool) (res []*item.Item) {
	for _, it := range storage.Content {
		if tester(it) {
			res = append(res, &it)
		}
	}
	return
}

//First gather first items matching requirements
func (storage *Storage) First(tester func(item.Item) bool) *item.Item {
	for _, it := range storage.Content {
		if tester(it) {
			return &it
		}
	}
	return nil
}

//Last gather last items matching requirements
func (storage *Storage) Last(tester func(item.Item) bool) (res *item.Item) {
	for _, it := range storage.Content {
		if tester(it) {
			res = &it
		}
	}
	return
}
