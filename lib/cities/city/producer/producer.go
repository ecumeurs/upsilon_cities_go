package producer

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"upsilon_cities_go/lib/cities/item"
	"upsilon_cities_go/lib/cities/storage"
	"upsilon_cities_go/lib/cities/tools"
)

type requirement struct {
	RessourceType string
	Quality       tools.IntRange
	Quantity      int
}

type upgradepoint struct {
	Total int
	Used  int
}

type upgradeHistory struct {
	QualityMin  int
	QualitMax   int
	QuantityMin int
	QuantityMax int
	History     []string
}

type bigUpgradeHistory struct {
	Delay       int // in cycles
	BasePrice   int
	QualityMin  int
	QualitMax   int
	QuantityMin int
	QuantityMax int
	History     []string
}

//Producer tell what it produce, within which criteria
type Producer struct {
	ID                int
	Name              string
	ProductName       string
	ProductType       string
	UpgradePoint      upgradepoint
	BigUpgradePoint   upgradepoint
	Quality           tools.IntRange
	Quantity          tools.IntRange
	UpgradeHistory    upgradeHistory
	BigUpgradeHistory bigUpgradeHistory
	BasePrice         int
	Requirements      []requirement
	Delay             int // in cycles
	Level             int // mostly informative, as levels will be applied directly to ranges, requirements and delay
	CurrentXP         int
	NextLevel         int
}

//Production active production stuff ;)
type Production struct {
	ProducerID  int
	StartTime   time.Time
	EndTime     time.Time
	Production  item.Item
	Reservation int64 // storage space reservation ticket

}

//Produce create a new item based on template
func (prod *Producer) produce() (res item.Item) {
	res.Name = prod.ProductName
	res.Type = prod.ProductType
	res.Quality = prod.Quality.Roll()
	res.Quantity = prod.Quantity.Roll()
	res.BasePrice = prod.BasePrice
	return
}

func (rq requirement) String() string {
	return fmt.Sprintf("%d x %s Q[%d-%d]", rq.Quantity, rq.RessourceType, rq.Quality.Min, rq.Quality.Max)
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
func (prod *Producer) Upgrade(action int, err error) {
	switch action {
	case 0: //Min Quantity: +1
		fmt.Println("one")
	case 1: //Max Quantity: +1
		fmt.Println("one")
	case 2: //Min Quality: +1
		fmt.Println("two")
	case 3: //Max Quality: +1
		fmt.Println("three")
	case 4: //Price: +1
		fmt.Println("one")
	case 5: //Delay: -1
		fmt.Println("two")
	case 6: //Min Quantity: +5
		fmt.Println("three")
	case 7: //Max Quantity: +5
		fmt.Println("one")
	case 8: //Min Quality: +5
		fmt.Println("two")
	case 9: //Max Quality: +5
		fmt.Println("three")
	}
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

//CanProduceShort tell whether it's able to produce item
func CanProduceShort(store *storage.Storage, prod *Producer) (producable bool, err error) {
	// producable immediately ?

	count := 0
	missing := make([]string, 0)
	found := make(map[string]int)
	var missitem string
	for _, v := range prod.Requirements {
		found[v.RessourceType] = 0
		count += v.Quantity

		for _, foundling := range store.All(storage.ByTypeNQuality(v.RessourceType, v.Quality)) {
			found[v.RessourceType] += foundling.Quantity
		}

		if found[v.RessourceType] < v.Quantity {
			missitem = fmt.Sprintf("%s need %d have %d", v.String(), v.Quantity, found[v.RessourceType])
			missing = append(missing, missitem)
		}
	}

	if len(missing) > 0 {
		return false, fmt.Errorf("not enough ressources: %s", strings.Join(missing, ", "))
	}

	if store.Spaceleft()+count < prod.Quantity.Min {
		return false, fmt.Errorf("not enough space available: potentially got: %d required %d", (store.Spaceleft() + count), prod.Quantity.Min)
	}
	return true, nil
}

//CanProduce tell whether it's able to produce item, if it can produce it relayabely or if not, tell why.
func CanProduce(store *storage.Storage, prod *Producer, ressourcesGenerators map[int]*Producer) (producable bool, nb int, recurrent bool, err error) {
	// producable immediately ?

	count := 0
	missing := make([]string, 0)
	found := make(map[string]int)
	for _, v := range prod.Requirements {
		found[v.RessourceType] = 0
		count += v.Quantity

		for _, foundling := range store.All(storage.ByTypeNQuality(v.RessourceType, v.Quality)) {
			found[v.RessourceType] += foundling.Quantity
		}

		if found[v.RessourceType] < v.Quantity {
			missing = append(missing, v.String())
		}
	}

	if len(missing) > 0 {
		return false, 0, false, fmt.Errorf("not enough ressources: %s", strings.Join(missing, ", "))
	}

	if store.Spaceleft()-count < prod.Quantity.Min {
		return false, 0, false, fmt.Errorf("not enough space available: potentially got: %d required %d", (store.Spaceleft() - count), prod.Quantity.Min)
	}

	producable = true
	// we can at least produce one this.

	available := make(map[string]int)
	// check if we can produce more than one

	for _, v := range prod.Requirements {
		available[v.RessourceType] = found[v.RessourceType] / v.Quantity
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
			if gen.ProductType == v.RessourceType {
				if gen.Quality.Min >= v.Quality.Min {
					found[v.RessourceType] += gen.Quantity.Min
				}
			}
		}
	}

	for _, v := range prod.Requirements {
		if _, has := found[v.RessourceType]; has {
			available[v.RessourceType] = found[v.RessourceType] / v.Quantity
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

		for _, foundling := range store.All(storage.ByTypeNQuality(v.RessourceType, v.Quality)) {
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
	producable, _ := CanProduceShort(store, prod)

	// reserve place for to be coming products...

	if !producable {
		return nil, errors.New("unable to use this Producer")
	}

	production := new(Production)

	production.StartTime = tools.RoundTime(startDate)
	production.EndTime = tools.AddCycles(production.StartTime, prod.Delay)
	production.ProducerID = prod.ID
	production.Production = prod.produce()

	err := deductProducFromStorage(store, prod)

	if err != nil {
		return nil, err
	}

	// from here we already know that space left >= Min production quantity.
	production.Production.Quantity = tools.Min(production.Production.Quantity, store.Spaceleft())
	production.Reservation, err = store.Reserve(production.Production.Quantity)

	if err != nil {
		return nil, err
	}

	return production, nil
}

//IsFinished tell whether production is finished or not ;)
func (prtion *Production) IsFinished(now time.Time) (finished bool) {
	finished = now.After(prtion.EndTime) || now.Equal(prtion.EndTime)
	log.Printf("Producer: %s : End: %s, Now %s, Finished ? %v", prtion.Production.Name, prtion.EndTime.Format(time.RFC3339), now.Format(time.RFC3339), finished)
	return
}

//ProductionCompleted Update store
func ProductionCompleted(store *storage.Storage, prtion *Production, nextUpdate time.Time) error {
	if prtion.IsFinished(nextUpdate) {
		return store.Claim(prtion.Reservation, prtion.Production)
	}
	return errors.New("unable to complete production (not finished)")
}
