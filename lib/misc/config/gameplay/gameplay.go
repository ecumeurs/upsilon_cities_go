package gameplay

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"upsilon_cities_go/lib/misc/config/system"
)

var configuration map[string]interface{}

//LoadConf read configuration file and fills local data.
func LoadConf() {
	f, ferr := os.Open(system.MakePath("config/gameplay.json"))
	if ferr != nil {
		log.Fatalf("GameplayConf: No Gameplay conf data file present: %s", ferr)
	}

	nameJSON, ferr := ioutil.ReadAll(f)
	if ferr != nil {
		log.Fatalf("GameplayConf: No Gameplay conf data file found but unable to read it all: %s", ferr)
	}

	f.Close()

	configuration = make(map[string]interface{})
	json.Unmarshal(nameJSON, &configuration)
}

//Get seeks value in configuration  for provided key
func Get(name string, def string) string {
	if configuration == nil {
		return def
	}
	if v, found := configuration[name]; found {
		return v.(string)
	}
	log.Printf("Gameplay: Attempting to use unknown rule: %s with default value: %s", name, def)
	return def
}

//GetFloat seeks value in configuration  for provided key
func GetFloat(name string, def float64) float64 {
	if configuration == nil {
		return def
	}
	if v, found := configuration[name]; found {
		return v.(float64)
	}
	log.Printf("Gameplay: Attempting to use unknown rule: %s with default value: %f", name, def)
	return def
}

//GetInt seeks value in configuration  for provided key
func GetInt(name string, def int) int {
	if configuration == nil {
		return def
	}
	if v, found := configuration[name]; found {
		return int(v.(float64))
	}
	log.Printf("Gameplay: Attempting to use unknown rule: %s with default value: %d", name, def)
	return def
}

//GetBool seeks value in configuration  for provided key
func GetBool(name string, def bool) bool {
	if configuration == nil {
		return def
	}
	if v, found := configuration[name]; found {
		return v.(bool)
	}
	log.Printf("Gameplay: Attempting to use unknown rule: %s with default value: %v", name, def)
	return def
}
