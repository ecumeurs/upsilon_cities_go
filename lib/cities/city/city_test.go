package city

import (
	"testing"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/misc/generator"
)

func prepare() {
	generator.CreateSampleFile()
	generator.Init()

	producer.CreateSampleFile()
	producer.Load()
}

func TestGeneratedCityHasDistinctRessources(t *testing.T) {
	prepare()
	// as it's random, do the check like hundred times ...
	for i := 0; i < 100; i++ {
		city := New()
		names := make(map[string]bool)
		for _, v := range city.RessourceProducers {
			if names[v.ProductName] {
				t.Errorf("Has already a ressource producer of same name")
				return
			} else {
				names[v.ProductName] = true
			}
		}
	}
}

func TestGeneratedCityHasDistinctFactories(t *testing.T) {
	prepare()
	// as it's random, do the check like hundred times ...
	for i := 0; i < 100; i++ {
		city := New()
		names := make(map[string]bool)
		for _, v := range city.ProductFactories {
			if names[v.ProductName] {
				t.Errorf("Has already a factory of same name")
				return
			} else {
				names[v.ProductName] = true
			}
		}
	}
}
