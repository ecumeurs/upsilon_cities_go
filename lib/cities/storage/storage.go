package storage

import (
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/tools"
)

//Storage contains every item of city and storage capacity
type Storage struct {
	Capacity int
	Content  []item.Item
}

//Count return the nb of item in Storage
func (storage *Storage) Count() int {
	var total int
	for _, item := range storage.Content {
		total += item.Quantity
	}

	return total
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
