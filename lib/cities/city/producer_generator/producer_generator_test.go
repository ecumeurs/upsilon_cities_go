package producer_generator

import (
	"encoding/json"
	"log"
	"testing"
	"upsilon_cities_go/lib/cities/city/producer"
)

func TestCanRegisterAResourceFactory(t *testing.T) {

	Initialize()
	baseID := 1

	{
		f := new(Factory)
		f.IsAdvanced = false
		f.Delay = 1
		f.IsRessource = true
		f.Requirements = make([]producer.Requirement, 0)
		f.ProducerName = "TestRessource"

		var p producer.Product
		p.ItemTypes = append(p.ItemTypes, "TestItemType2")
		p.ItemName = "TestItemType2 ?!"
		p.Quality.Min = 20
		p.Quality.Max = 25
		p.Quantity.Min = 1
		p.Quantity.Max = 2
		f.Products = append(f.Products, p)

		loadFactory(f, baseID, "")
		baseID++
	}

	validate()

}

func TestCanRegisterAProductFactory(t *testing.T) {

	Initialize()
	baseID := 1

	{
		f := new(Factory)
		f.IsAdvanced = false
		f.Delay = 1
		f.IsRessource = false
		f.Requirements = make([]producer.Requirement, 0)
		f.ProducerName = "TestMaker"

		var p producer.Product
		p.ItemTypes = append(p.ItemTypes, "ItemType3")
		p.ItemName = "TestItem"
		p.Quality.Min = 20
		p.Quality.Max = 25
		p.Quantity.Min = 1
		p.Quantity.Max = 2
		f.Products = append(f.Products, p)

		var r producer.Requirement
		r.ItemTypes = []string{"TestItemType2"}
		r.Denomination = "Some item"
		r.Quantity = 3
		r.Quality.Min = 5
		r.Quality.Max = 50
		f.Requirements = append(f.Requirements, r)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		f.IsAdvanced = false
		f.Delay = 1
		f.IsRessource = true
		f.Requirements = make([]producer.Requirement, 0)
		f.ProducerName = "TestRessource"

		var p producer.Product
		p.ItemTypes = append(p.ItemTypes, "TestItemType2")
		p.ItemName = "TestItemType2 ?!"
		p.Quality.Min = 20
		p.Quality.Max = 25
		p.Quantity.Min = 1
		p.Quantity.Max = 2
		f.Products = append(f.Products, p)

		loadFactory(f, baseID, "")
		baseID++
	}
	validate()

}

func TestCanRegisterAProductFactoryWithComplexResource(t *testing.T) {

	Initialize()
	baseID := 1

	{
		f := new(Factory)
		f.IsAdvanced = false
		f.Delay = 1
		f.IsRessource = false
		f.Requirements = make([]producer.Requirement, 0)
		f.ProducerName = "TestMaker"

		var p producer.Product
		p.ItemTypes = append(p.ItemTypes, "ItemType3")
		p.ItemName = "TestItem"
		p.Quality.Min = 20
		p.Quality.Max = 25
		p.Quantity.Min = 1
		p.Quantity.Max = 2
		f.Products = append(f.Products, p)

		var r producer.Requirement
		r.ItemTypes = []string{"TestItemType2"}
		r.Denomination = "Some item"
		r.Quantity = 3
		r.Quality.Min = 5
		r.Quality.Max = 50
		f.Requirements = append(f.Requirements, r)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		f.IsAdvanced = false
		f.Delay = 1
		f.IsRessource = true
		f.Requirements = make([]producer.Requirement, 0)
		f.ProducerName = "TestRessource"

		var p producer.Product
		p.ItemTypes = append(p.ItemTypes, "TestItemType2")
		p.ItemTypes = append(p.ItemTypes, "OtherResourceType")
		p.ItemName = "Complex TestItemType2 ?!"
		p.Quality.Min = 20
		p.Quality.Max = 25
		p.Quantity.Min = 1
		p.Quantity.Max = 2
		f.Products = append(f.Products, p)

		loadFactory(f, baseID, "")
		baseID++
	}
	validate()
}

func TestWhenMultiplesRessourcesAreAvailableValidateWithTheMostAppropriate(t *testing.T) {

	// Point here isn't about json loading
	// But when 2 factories match requirements for a producer
	// Choose for validation the one with the least requirements ;)

	Initialize()

	baseID := 1

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": false,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
				],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": true,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}
	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
			],
			"Products": [
				{
					"ItemTypes": [
						"Lingot", "Fer", "Metal"
					],
					"ItemName": "Lingot de Fer",
					"Quality": {
						"Min": 15,
						"Max": 25
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 30
				},
				{
					"ItemTypes": [
						"Poudre", "Soufre", "Mineral"
					],
					"ItemName": "Poudre de Soufre",
					"Quality": {
						"Min": 12,
						"Max": 28
					},
					"Quantity": {
						"Min": 0,
						"Max": 1
					},
					"BasePrice": 6
				}
			],
			"Delay": 3,
			"IsAdvanced": false,
			"ProducerName": "Fonderie"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}
	validate()
}

func TestProducerRequiringTypes(t *testing.T) {

	Initialize()

	activeResources := []string{"Bois"}

	baseID := 1
	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
		"Requirements": [
			{
				"ItemTypes": [
					"Bois"
				],
				"ItemName": "Tronc",
				"Quality": {
					"Min": 0,
					"Max": 150
				},
				"Quantity": 5
			}
		],
		"Products": [
			{
				"ItemTypes": [
					"Planche de bois"
				],
				"ItemName": "Planche de bois",
				"Quality": {
					"Min": 15,
					"Max": 25
				},
				"Quantity": {
					"Min": 1,
					"Max": 1
				},
				"BasePrice": 30
			}
		],
		"Delay": 3,
		"IsAdvanced": false,
		"ProducerName": "Scierie"
	}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": false,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
				],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": true,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}
	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
			],
			"Products": [
				{
					"ItemTypes": [
						"Lingot", "Fer", "Metal"
					],
					"ItemName": "Lingot de Fer",
					"Quality": {
						"Min": 15,
						"Max": 25
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 30
				},
				{
					"ItemTypes": [
						"Poudre", "Soufre", "Mineral"
					],
					"ItemName": "Poudre de Soufre",
					"Quality": {
						"Min": 12,
						"Max": 28
					},
					"Quantity": {
						"Min": 0,
						"Max": 1
					},
					"BasePrice": 6
				}
			],
			"Delay": 3,
			"IsAdvanced": false,
			"ProducerName": "Fonderie"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	candidateFactories := ProducerRequiringTypes(activeResources, true)

	var fact *Factory
	if len(candidateFactories) != 1 {
		t.Errorf("Expected to have one candidate, got :%d", len(candidateFactories))
		for _, v := range candidateFactories {
			log.Printf("Factory: %v", v)
		}
		return
	}

	fact = candidateFactories[0]

	if fact.ID != 1 {
		t.Error("Expected found factory to be factory ided 1")
	}
}

func TestProducerRequiringTypesNotExclusive(t *testing.T) {

	Initialize()
	activeResources := []string{"Bois"}

	baseID := 1
	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
		"Requirements": [
			{
				"ItemTypes": [
					"Bois"
				],
				"ItemName": "Tronc",
				"Quality": {
					"Min": 0,
					"Max": 150
				},
				"Quantity": 5
			}
		],
		"Products": [
			{
				"ItemTypes": [
					"Planche de bois"
				],
				"ItemName": "Planche de bois",
				"Quality": {
					"Min": 15,
					"Max": 25
				},
				"Quantity": {
					"Min": 1,
					"Max": 1
				},
				"BasePrice": 30
			}
		],
		"Delay": 3,
		"IsAdvanced": false,
		"ProducerName": "Scierie"
	}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": false,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
				],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": true,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}
	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
			],
			"Products": [
				{
					"ItemTypes": [
						"Lingot", "Fer", "Metal"
					],
					"ItemName": "Lingot de Fer",
					"Quality": {
						"Min": 15,
						"Max": 25
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 30
				},
				{
					"ItemTypes": [
						"Poudre", "Soufre", "Mineral"
					],
					"ItemName": "Poudre de Soufre",
					"Quality": {
						"Min": 12,
						"Max": 28
					},
					"Quantity": {
						"Min": 0,
						"Max": 1
					},
					"BasePrice": 6
				}
			],
			"Delay": 3,
			"IsAdvanced": false,
			"ProducerName": "Fonderie"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	candidateFactories := ProducerRequiringTypes(activeResources, false)

	var fact *Factory
	if len(candidateFactories) != 2 {
		t.Errorf("Expected to have two candidate, got :%d", len(candidateFactories))
		for _, v := range candidateFactories {
			log.Printf("Factory: %v", v)
		}
		return
	}

	fact = candidateFactories[0]

	if fact.ID != 1 {
		t.Error("Expected found factory to be factory ided 1")
	}

	fact = candidateFactories[1]
	if fact.ID != 2 {
		t.Error("Expected found second factory to be factory ided 2")
	}
}

func TestResourceProducerProducingTypes(t *testing.T) {

	Initialize()

	activeResources := []string{"Bois"}

	baseID := 1
	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
		"Requirements": [
		],
		"Products": [
			{
				"ItemTypes": [
					"Bois"
				],
				"ItemName": "Tronc d'arbre",
				"Quality": {
					"Min": 15,
					"Max": 25
				},
				"Quantity": {
					"Min": 1,
					"Max": 1
				},
				"BasePrice": 30
			}
		],
		"Delay": 3,
		"IsAdvanced": false,
		"ProducerName": "Sylviculteur"
	}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": false,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
				],
			"Products": [
				{
					"ItemTypes": [
						"Minerai", "Fer", "Metal"
					],
					"ItemName": "Minerai de Fer",
					"Quality": {
						"Min": 10,
						"Max": 20
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 5
				}
			],
			"Delay": 1,
			"IsAdvanced": true,
			"ProducerName": "Mine de Fer"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}
	{
		f := new(Factory)
		json.Unmarshal([]byte(`{
			"Requirements": [
				{
					"ItemTypes": [
						"Minerai", "Fer"
					],
					"ItemName": "",
					"Quality": {
						"Min": 0,
						"Max": 150
					},
					"Quantity": 5,
					"Denomination": "Minerai de Fer"
				}
			],
			"Products": [
				{
					"ItemTypes": [
						"Lingot", "Fer", "Metal"
					],
					"ItemName": "Lingot de Fer",
					"Quality": {
						"Min": 15,
						"Max": 25
					},
					"Quantity": {
						"Min": 1,
						"Max": 1
					},
					"BasePrice": 30
				},
				{
					"ItemTypes": [
						"Poudre", "Soufre", "Mineral"
					],
					"ItemName": "Poudre de Soufre",
					"Quality": {
						"Min": 12,
						"Max": 28
					},
					"Quantity": {
						"Min": 0,
						"Max": 1
					},
					"BasePrice": 6
				}
			],
			"Delay": 3,
			"IsAdvanced": false,
			"ProducerName": "Fonderie"
		}`), &f)

		loadFactory(f, baseID, "")
		baseID++
	}

	candidateFactories := ResourceProducerProducingTypes(activeResources)

	var fact *Factory
	if len(candidateFactories) != 1 {
		t.Errorf("Expected to have one candidate, got :%d", len(candidateFactories))
		for _, v := range candidateFactories {
			log.Printf("Factory: %v", v)
		}
		return
	}

	fact = candidateFactories[0]

	if fact.ID != 1 {
		t.Error("Expected found factory to be factory ided 1")
	}
}
