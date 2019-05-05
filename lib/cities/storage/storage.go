package storage

import "upsilon_cities_go/lib/cities/item"

//Storage contains every item of city and storage capacity
type Storage struct {
	Capacity int
	Content  []item.Item
}

//Isfull return a boolean if nb item reach capacity
func (storage *Storage) Isfull() bool {
	var total int
	for _, item := range storage.Content {
		total += item.Quantity
	}

	return total == storage.Capacity
}

//Spaceleft return space left depending of capacity
func (storage *Storage) Spaceleft() int {
	var total int
	for _, item := range storage.Content {
		total += item.Quantity
	}

	return storage.Capacity - total
}
