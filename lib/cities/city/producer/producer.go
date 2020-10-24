package producer

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
)

const (
	quantityOne  int = 0
	qualityOne   int = 1
	delay        int = 2
	quantityFive int = 3
	qualityFive  int = 4
)

//Requirement item types or name to required to instantiate a product.
type Requirement struct {
	ItemTypes    []string
	ItemName     string
	Quality      tools.IntRange
	Quantity     int
	Denomination string
}

//Product description
type Product struct {
	ID          int
	ItemTypes   []string
	ItemName    string
	Quality     tools.IntRange
	Quantity    tools.IntRange
	BasePrice   int
	UpgradeInfo upgrade
}

func (p Product) String() string {
	return fmt.Sprintf("(%s [%s]) x %d", p.ItemName, strings.Join(p.ItemTypes, ","), p.Quantity.Min)
}

//StringShort short version of a product
func (p Product) StringShort() string {
	return fmt.Sprintf("%s [%s]", p.ItemName, strings.Join(p.ItemTypes, ","))
}

type upgradepoint struct {
	Total int
	Used  int
}

type upgrade struct {
	Quality  int
	Quantity int
}

//Producer tell what it produce, within which criteria
type Producer struct {
	ID              int
	FactoryID       int `json:"-"` // for general identification (uniqueness)
	Name            string
	UpgradePoint    upgradepoint
	BigUpgradePoint upgradepoint
	History         map[int][]int
	Requirements    []Requirement
	Products        map[int]Product
	Delay           int // in cycles
	Level           int // mostly informative, as levels will be applied directly to ranges, requirements and delay
	CurrentXP       int
	NextLevel       int
	Advanced        bool
	UpgradeDelay    int // Delay upgrade
	LastActivity    time.Time
}

//Production active production stuff ;)
type Production struct {
	ProducerID   int
	ProducerName string
	StartTime    time.Time
	EndTime      time.Time
	Production   []item.Item
	Reservation  int64 // storage space reservation ticket

}

//Produce create a new item based on template
func (prod *Producer) produce() (res []item.Item) {
	for _, v := range prod.Products {
		var rs item.Item
		rs.Name = v.ItemName
		rs.Type = v.ItemTypes
		rs.Quality = v.GetQuality().Roll()
		rs.Quantity = v.GetQuantity().Roll()
		rs.BasePrice = v.BasePrice
		res = append(res, rs)
	}
	prod.LastActivity = tools.RoundNow()
	return
}

//GetDelay Get Delay with Upgrade
func (prod *Producer) GetDelay() int {
	return prod.Delay * ((100.00 - prod.UpgradeDelay) / 100.00)
}

//GetQuality Get Quality with Upgrade
func (p Product) GetQuality() tools.IntRange {
	min := p.Quality.Min + p.Quality.Min*(p.UpgradeInfo.Quality/100)
	max := p.Quality.Max + p.Quality.Max*(p.UpgradeInfo.Quality/100)
	return tools.IntRange{Min: min, Max: max}
}

//GetQuantity Get Quantity with Upgrade
func (p Product) GetQuantity() tools.IntRange {
	min := p.Quantity.Min + p.Quantity.Min*(p.UpgradeInfo.Quantity/100)
	max := p.Quantity.Max + p.Quantity.Max*(p.UpgradeInfo.Quantity/100)
	return tools.IntRange{Min: min, Max: max}
}

func (rq Requirement) String() string {
	rsc := ""
	if len(rq.ItemTypes) > 0 {
		rsc = fmt.Sprintf("(%s)", strings.Join(rq.ItemTypes, ","))
	} else {
		rsc = rq.ItemName
	}

	return fmt.Sprintf("%d x %s Q[%d-%d]", rq.Quantity, rsc, rq.Quality.Min, rq.Quality.Max)
}

//Leveling all leveling related action
func (prod *Producer) Leveling(point int) {
	prod.CurrentXP += point
	for prod.CurrentXP >= prod.NextLevel {
		prod.CurrentXP = prod.CurrentXP - prod.NextLevel
		prod.Level++
		prod.UpgradePoint.Total++
		if prod.Level%5 == 0 {
			prod.BigUpgradePoint.Total++
		}
		prod.NextLevel = GetNextLevel(prod.Level)
	}
}

//Upgrade Upgrade producer depending of action
func (prod *Producer) Upgrade(action int, productID int) (result bool) {
	p, _ := prod.Products[productID]

	var canUpgrade bool
	var used *int

	switch action {
	case quantityOne, qualityOne: //Min Quantity: +1
		canUpgrade = prod.CanUpgrade()
		used = &prod.UpgradePoint.Used
	case delay, quantityFive, qualityFive: //Max Quantity: +1
		canUpgrade = prod.CanBigUpgrade()
		used = &prod.BigUpgradePoint.Used
	}

	if canUpgrade {

		switch action {
		case quantityOne: //Min Quantity: +1
			p.UpgradeInfo.Quantity++

		case qualityOne: //Min Quality: +1
			p.UpgradeInfo.Quality++

		case quantityFive: //Min Quantity: +5
			p.UpgradeInfo.Quantity += 5

		case qualityFive: //Min Quality: +5
			p.UpgradeInfo.Quality += 5

		case delay: //Delay: +1
			prod.UpgradeDelay++
		}

		*used++
		prod.Products[productID] = p
		prod.History[productID] = append(prod.History[productID], action) //Product ID 0 for producer
	}

	return canUpgrade
}

//CanUpgrade Producer can make a simple Upgrade
func (prod *Producer) CanUpgrade() bool {
	return (prod.UpgradePoint.Total - prod.UpgradePoint.Used) > 0
}

//CanBigUpgrade Producer can make a big upgrade
func (prod *Producer) CanBigUpgrade() bool {
	return (prod.BigUpgradePoint.Total - prod.BigUpgradePoint.Used) > 0
}

//GetNextLevel return next level needed xp
func GetNextLevel(acLevel int) int {
	return 10 * acLevel
}

func seekItemsByRequirement(rq Requirement, store *storage.Storage) []item.Item {
	if len(rq.ItemTypes) > 0 {
		return store.All(storage.ByTypesNQuality(rq.ItemTypes, rq.Quality))
	}
	return store.All(storage.ByNameNQuality(rq.ItemName, rq.Quality))
}

//CanProduceShort tell whether it's able to produce item
func CanProduceShort(store *storage.Storage, prod *Producer) (producable bool, space bool, err error) {
	// producable immediately ?

	count := 0
	missing := make([]string, 0)
	found := make(map[string]int)
	var missitem string
	for _, v := range prod.Requirements {
		found[v.Denomination] = 0
		count += v.Quantity

		for _, foundling := range seekItemsByRequirement(v, store) {
			found[v.Denomination] += foundling.Quantity
		}

		if found[v.Denomination] < v.Quantity {
			missitem = fmt.Sprintf("%s need %d have %d", v.String(), v.Quantity, found[v.Denomination])
			missing = append(missing, missitem)
		}
	}

	if len(missing) > 0 {
		return false, false, fmt.Errorf("not enough ressources: %s", strings.Join(missing, ", "))
	}

	minProduced := 0
	for _, v := range prod.Products {
		minProduced += v.Quantity.Min
	}

	if store.Spaceleft()+count < minProduced {
		return false, true, fmt.Errorf("not enough space available: potentially got: %d required %d", (store.Spaceleft() + count), minProduced)
	}
	return true, false, nil
}

//CanProduce tell whether it's able to produce item, if it can produce it relayabely or if not, tell why.
func CanProduce(store *storage.Storage, prod *Producer, ressourcesGenerators map[int]*Producer) (producable bool, nb int, recurrent bool, err error) {
	// producable immediately ?

	count := 0
	missing := make([]string, 0)
	found := make(map[string]int)
	for _, v := range prod.Requirements {
		found[v.Denomination] = 0
		count += v.Quantity

		for _, foundling := range seekItemsByRequirement(v, store) {
			found[v.Denomination] += foundling.Quantity
		}

		if found[v.Denomination] < v.Quantity {
			missing = append(missing, v.String())
		}
	}

	if len(missing) > 0 {
		return false, 0, false, fmt.Errorf("not enough ressources: %s", strings.Join(missing, ", "))
	}

	minProduced := 0
	for _, v := range prod.Products {
		minProduced += v.Quantity.Min
	}

	if store.Spaceleft()-count < minProduced {
		return false, 0, false, fmt.Errorf("not enough space available: potentially got: %d required %d", (store.Spaceleft() - count), minProduced)
	}

	producable = true
	// we can at least produce one this.

	available := make(map[string]int)
	// check if we can produce more than one

	for _, v := range prod.Requirements {
		available[v.Denomination] = found[v.Denomination] / v.Quantity
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
			for _, p := range gen.Products {
				if len(v.ItemTypes) > 0 && tools.ListInStringList(v.ItemTypes, p.ItemTypes) {
					if p.Quality.Min >= v.Quality.Min {
						found[v.Denomination] += p.Quantity.Min
					}
				} else if v.ItemName == p.ItemName {
					if p.Quality.Min >= v.Quality.Min {
						found[v.Denomination] += p.Quantity.Min
					}
				}
			}
		}
	}

	for _, v := range prod.Requirements {
		if found[v.Denomination] > 0 {
			available[v.Denomination] = found[v.Denomination] / v.Quantity
		} else {
			return producable, nb, false, nil
		}
	}

	return producable, nb, true, nil
}

//DeductProducFromStorage attempt to remove necessary items from store to start producer.
func deductProducFromStorage(store *storage.Storage, prod *Producer) error {
	found := make(map[int64]int)
	for _, v := range prod.Requirements {
		target := v.Quantity

		for _, foundling := range seekItemsByRequirement(v, store) {
			used := tools.Min(target, foundling.Quantity)
			found[foundling.ID] = used
			target -= used

			if target <= 0 {
				break
			}
		}

		if target > 0 {
			return fmt.Errorf("Unable to fit requirement %s", v.String())
		}
	}

	for k, v := range found {
		store.Remove(k, v)
	}
	return nil
}

//Produce Kicks in Producer and instantiate a Production, if able.
func Produce(store *storage.Storage, prod *Producer, startDate time.Time) (*Production, error) {
	producable, _, _ := CanProduceShort(store, prod)

	// reserve place for to be coming products...

	if !producable {
		return nil, errors.New("unable to use this Producer")
	}

	production := new(Production)

	production.StartTime = tools.RoundTime(startDate)
	production.EndTime = tools.AddCycles(production.StartTime, prod.GetDelay())
	production.ProducerID = prod.ID
	production.ProducerName = prod.Name
	production.Production = prod.produce()

	err := deductProducFromStorage(store, prod)

	if err != nil {
		return nil, err
	}

	// from here we already know that space left >= Min production quantity.

	totalQty := 0
	for _, v := range production.Production {
		totalQty += v.Quantity
	}

	ntotalQty := 0
	if totalQty > store.Spaceleft() {
		// compute prorata per item and apply to space left.
		for idx, v := range production.Production {
			v.Quantity = int(math.Round(float64(v.Quantity) * (float64(v.Quantity) / float64(totalQty))))
			ntotalQty += v.Quantity
			production.Production[idx] = v
		}
	} else {
		ntotalQty = totalQty
	}

	production.Reservation, err = store.Reserve(tools.Min(ntotalQty, store.Spaceleft()))

	if err != nil {
		return nil, err
	}

	return production, nil
}

//IsFinished tell whether production is finished or not ;)
func (prtion *Production) IsFinished(now time.Time) (finished bool) {
	finished = now.After(prtion.EndTime) || now.Equal(prtion.EndTime)
	log.Printf("Producer: %d %s : End: %s, Now %s, Finished ? %v", prtion.ProducerID, prtion.ProducerName, prtion.EndTime.Format(time.RFC3339), now.Format(time.RFC3339), finished)
	return
}

//ProductionCompleted Update store
func ProductionCompleted(store *storage.Storage, prtion *Production, nextUpdate time.Time) error {
	if prtion.IsFinished(nextUpdate) {
		return store.Claim(prtion.Reservation, prtion.Production)
	}
	return errors.New("unable to complete production (not finished)")
}
