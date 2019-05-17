package storage

import (
	"errors"
	"fmt"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/tools"
)

//Storage contains every item of city and storage capacity. Note `json:"-"` means it won't be exported as json ...
type Storage struct {
	Capacity     int
	Content      map[int64]item.Item
	CurrentMaxID int64         `json:"-"`
	Reservations map[int64]int `json:"-"` // reservation_id > reserved size.
}

//New storage !
func New() (res *Storage) {
	res = new(Storage)
	res.Content = make(map[int64]item.Item)
	res.Reservations = make(map[int64]int)
	res.CurrentMaxID = 1
	res.Capacity = 10
	return
}

//Get seek out an item in storage.
func (storage *Storage) Get(ID int64) (res item.Item, found bool) {
	res, found = storage.Content[ID]
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
	for _, sz := range storage.Reservations {
		total += sz
	}

	return total
}

//Spaceleft return space left depending of capacity
func (storage *Storage) Spaceleft() int {
	return storage.Capacity - storage.Count()
}

//Add item to storage
func (storage *Storage) Add(it item.Item) error {
	if storage.Spaceleft() < it.Quantity {
		return errors.New("unable to insert item, no space left")
	}

	storedit, err := storage.First(ByMatch(it))

	if err == nil {
		storedit.Quantity += it.Quantity
		storage.Content[storedit.ID] = storedit
		return nil
	}

	it.ID = storage.CurrentMaxID
	storage.Content[storage.CurrentMaxID] = it
	storage.CurrentMaxID++

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

	if itm.Quantity == 0 {
		delete(storage.Content, id)
	}
	return nil
}

//Isfull return a boolean if nb item reach capacity
func (storage *Storage) Isfull() bool {
	return storage.Count() == storage.Capacity
}

//Has tell whether store has item requested in number.
func (storage *Storage) Has(itType string, itNb int) bool {
	for _, it := range storage.Content {
		if tools.InStringList(itType, it.Type) {
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
		if tools.InStringList(itType, it.Type) {
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
		if tools.InStringList(itType, it.Type) {
			if tester(it) {
				return true
			}
		}
	}
	return false
}

//ByMatch function generator select item by match.
func ByMatch(lhs item.Item) func(item.Item) bool {
	return func(i item.Item) bool {
		return lhs.Match(i)
	}
}

//ByType function generator select item by type.
func ByType(itype string) func(item.Item) bool {
	return func(i item.Item) bool {
		return tools.InStringList(itype, i.Type)
	}
}

//ByTypeNQuality function generator select item by type and within quality range.
func ByTypeNQuality(itype string, ql tools.IntRange) func(item.Item) bool {
	return func(i item.Item) bool {
		return tools.InStringList(itype, i.Type) && tools.InEqRange(i.Quality, ql)
	}
}

//ByTypeOrNameNQuality function generator select item by type and within quality range.
func ByTypeOrNameNQuality(itype string, tpe bool, ql tools.IntRange) func(item.Item) bool {
	return func(i item.Item) bool {
		return ((tpe && tools.InStringList(itype, i.Type)) || (!tpe && i.Name == itype)) && tools.InEqRange(i.Quality, ql)
	}
}

//All gather all items matching requirements
func (storage *Storage) All(tester func(item.Item) bool) (res []item.Item) {
	for _, it := range storage.Content {
		if tester(it) {
			tmp := it
			res = append(res, tmp)
		}
	}
	return
}

//First gather first items matching requirements
func (storage *Storage) First(tester func(item.Item) bool) (item.Item, error) {
	for _, it := range storage.Content {
		if tester(it) {
			return it, nil
		}
	}
	return item.Item{}, errors.New("found no match")
}

//Last gather last items matching requirements
func (storage *Storage) Last(tester func(item.Item) bool) (res item.Item, err error) {
	err = errors.New("found no match")
	for _, it := range storage.Content {
		if tester(it) {
			res = it
			err = nil
		}
	}
	return
}

//Pretty provide a pretty display of the storage.
func (storage *Storage) Pretty() (res string) {
	for _, v := range storage.Content {
		res += fmt.Sprintf("\t%s\n", v.Pretty())
	}
	return
}

//Reserve space for futur usage.
func (storage *Storage) Reserve(size int) (id int64, err error) {
	if size < storage.Spaceleft() {
		id = storage.CurrentMaxID
		storage.CurrentMaxID++
		storage.Reservations[id] = size
		return id, nil
	}
	return 0, errors.New("unable to reserve space, not enough available")
}

//Claim reserved space
func (storage *Storage) Claim(id int64, it item.Item) (err error) {
	size, found := storage.Reservations[id]
	if !found {
		return errors.New("unable to claim space as provided identifier is unknown")
	}
	if size < it.Quantity {
		return errors.New("unable to insert item in storage as reserved capacity doesn't match provided item quantity")
	}

	delete(storage.Reservations, id)

	return storage.Add(it)
}

//GiveBack releases reserved space for public use.
func (storage *Storage) GiveBack(id int64) (err error) {
	_, found := storage.Reservations[id]
	if !found {
		return errors.New("unable to give back space as provided identifier is unknown")
	}

	delete(storage.Reservations, id)

	return nil
}
