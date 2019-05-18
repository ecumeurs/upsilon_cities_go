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
	quantityMinOne  int = 0
	quantityMaxOne  int = 1
	qualityMinOne   int = 2
	qualityMaxOne   int = 3
	basePrice       int = 4
	delay           int = 5
	quantityMinFive int = 6
	quantityMaxFive int = 7
	qualityMinFive  int = 8
	qualityMaxFive  int = 9
)

type requirement struct {
	ItemTypes    []string
	ItemName     string
	Quality      tools.IntRange
	Quantity     int
	Denomination string
}

type product struct {
	ID             int `json:"-"`
	ItemTypes      []string
	ItemName       string
	Quality        tools.IntRange
	Quantity       tools.IntRange
	BasePrice      int
	UpgradeInfo    upgrade    `json:"-"`
	BigUpgradeInfo bigUpgrade `json:"-"`
}

func (p product) String() string {
	return fmt.Sprintf("(%s [%s]) x %d", p.ItemName, strings.Join(p.ItemTypes, ","), p.Quantity.Min)
}

type upgradepoint struct {
	Total int
	Used  int
}

type upgrade struct {
	QualityMin  int
	QualityMax  int
	QuantityMin int
	QuantityMax int
}

type bigUpgrade struct {
	Delay       int // in cycles
	BasePrice   int
	QualityMin  int
	QualityMax  int
	QuantityMin int
	QuantityMax int
}

//Producer tell what it produce, within which criteria
type Producer struct {
	ID              int
	FactoryID       int `json:"-"` // for general identification (uniqueness)
	Name            string
	UpgradePoint    upgradepoint
	BigUpgradePoint upgradepoint
	History         []int
	Requirements    []requirement
	Products        map[int]product
	Delay           int // in cycles
	Level           int // mostly informative, as levels will be applied directly to ranges, requirements and delay
	CurrentXP       int
	NextLevel       int
	Advanced        bool
	BigUpgradeInfo  bigUpgrade
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
		rs.BasePrice = v.GetBasePrice()
		res = append(res, rs)
	}
	return
}

//GetDelay Get Delay with Upgrade
func (prod *Producer) GetDelay() int {
	return prod.Delay * ((100.00 - prod.BigUpgradeInfo.Delay) / 100.00)
}

//GetBasePrice Get BasePrice with Upgrade
func (prod product) GetBasePrice() int {
	return prod.BasePrice * ((100.00 + prod.BigUpgradeInfo.BasePrice) / 100.00)
}

//GetQuality Get Quality with Upgrade
func (prod product) GetQuality() tools.IntRange {
	min := (prod.Quality.Min + (prod.BigUpgradeInfo.QualityMin * 5) + prod.UpgradeInfo.QualityMin)
	max := (prod.Quality.Max + (prod.BigUpgradeInfo.QualityMax * 5) + prod.UpgradeInfo.QualityMax)
	return tools.IntRange{Min: min, Max: max}
}

//GetQuantity Get Quantity with Upgrade
func (prod product) GetQuantity() tools.IntRange {
	min := (prod.Quantity.Min + (prod.BigUpgradeInfo.QuantityMin * 5) + prod.UpgradeInfo.QuantityMin)
	max := (prod.Quantity.Max + (prod.BigUpgradeInfo.QuantityMax * 5) + prod.UpgradeInfo.QuantityMax)
	return tools.IntRange{Min: min, Max: max}
}

func (rq requirement) String() string {
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

	if action == delay {
		canUpgrade := prod.CanBigUpgrade()
		used := &prod.BigUpgradePoint.Used
		if canUpgrade {
			prod.BigUpgradeInfo.Delay++
			*used++
			prod.History = append(prod.History, action)
		}
		return canUpgrade
	}

	p, _ := prod.Products[productID]

	var stat *int
	var canUpgrade bool
	var used *int

	switch action {
	case 0, 1, 2, 3: //Min Quantity: +1
		canUpgrade = prod.CanUpgrade()
		used = &prod.UpgradePoint.Used
	case 4, 5, 6, 7, 8, 9: //Max Quantity: +1
		canUpgrade = prod.CanBigUpgrade()
		used = &prod.BigUpgradePoint.Used
	}

	switch action {
	case quantityMinOne: //Min Quantity: +1
		stat = &p.UpgradeInfo.QuantityMin

	case quantityMaxOne: //Max Quantity: +1
		stat = &p.UpgradeInfo.QuantityMax

	case qualityMinOne: //Min Quality: +1
		stat = &p.UpgradeInfo.QualityMin

	case qualityMaxOne: //Max Quality: +1
		stat = &p.UpgradeInfo.QualityMax

	case basePrice: //Price: +1
		stat = &p.BigUpgradeInfo.BasePrice

	case quantityMinFive: //Min Quantity: +5
		stat = &p.BigUpgradeInfo.QuantityMin

	case quantityMaxFive: //Max Quantity: +5
		stat = &p.BigUpgradeInfo.QuantityMax

	case qualityMinFive: //Min Quality: +5
		stat = &p.BigUpgradeInfo.QualityMin

	case qualityMaxFive: //Max Quality: +5
		stat = &p.BigUpgradeInfo.QualityMax
	}

	if canUpgrade {
		*stat++
		*used++
		prod.History = append(prod.History, action)
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

func seekItemsByRequirement(rq requirement, store *storage.Storage) []item.Item {
	if len(rq.ItemTypes) > 0 {
		return store.All(storage.ByTypesNQuality(rq.ItemTypes, rq.Quality))
	} else {
		return store.All(storage.ByNameNQuality(rq.ItemName, rq.Quality))
	}
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

//Product Kicks in Producer and instantiate a Production, if able.
func Product(store *storage.Storage, prod *Producer, startDate time.Time) (*Production, error) {
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
