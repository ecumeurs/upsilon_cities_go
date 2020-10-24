package city

import (
	"testing"
	"upsilon_cities_go/lib/cities/city/producer_generator"
	"upsilon_cities_go/lib/misc/generator"
)

func prepare() {
	generator.CreateSampleFile()
	generator.Load()

	producer_generator.CreateSampleFile()
	producer_generator.Load()
}

func TestCanGenerateACity(t *testing.T) {
	prepare()
	// as it's random, do the check like hundred times ...
	New()
}
func TestGeneratedCityHasDistinctRessources(t *testing.T) {
	prepare()
	// as it's random, do the check like hundred times ...
	for i := 0; i < 100; i++ {
		city := New()
		names := make(map[int]bool)
		for _, v := range city.RessourceProducers {
			if names[v.FactoryID] {
				t.Errorf("Has already a ressource producer of same id %d", v.FactoryID)
				return
			}
			names[v.FactoryID] = true
		}
	}
}

func TestGeneratedCityHasDistinctFactories(t *testing.T) {
	prepare()
	// as it's random, do the check like hundred times ...
	for i := 0; i < 100; i++ {
		city := New()
		names := make(map[int]bool)
		for _, v := range city.ProductFactories {
			if names[v.FactoryID] {
				t.Errorf("Has already a factory of same id %d", v.FactoryID)
				return
			}
			names[v.FactoryID] = true

		}
	}
}
