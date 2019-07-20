package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var configuration map[string]interface{}

//LoadConf read configuration file and fills local data.
func LoadConf() {
	f, ferr := os.Open(MakePath("config/system.json"))
	if ferr != nil {
		log.Fatalf("SystemConf: No System conf data file present: %s", ferr)
	}

	nameJSON, ferr := ioutil.ReadAll(f)
	if ferr != nil {
		log.Fatalf("SystemConf: No System conf data file found but unable to read it all: %s", ferr)
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
	return def
}

//GetFloat seeks value in configuration  for provided key
func GetFloat(name string, def float32) float32 {
	if configuration == nil {
		return def
	}
	if v, found := configuration[name]; found {
		return v.(float32)
	}
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
	return def
}

var root string
var rootSlash string

//Root seek root path of the project.
func Root() string {
	if root == "" {
		if GetBool("force_sys_root", false) {
			if strings.ContainsAny(Get("sys_root", ""), "\\") {
				root = Get("sys_root", "")
				rootSlash = filepath.ToSlash(Get("sys_root", ""))
			} else {
				root = Get("sys_root", "")
				rootSlash = Get("sys_root", "")
			}
			return root
		}

		// get working directory and go up until finding upsilon_cities_go
		dir, _ := os.Getwd()
		dir = filepath.ToSlash(dir)
		subs := strings.Split(dir, "/")

		// keep what's necessary up until last inserted is upsilon_cities_go
		nsubs := make([]string, 0)

		for _, v := range subs {
			nsubs = append(nsubs, v)
			if v == "upsilon_cities_go" {
				break
			}

		}

		rootSlash = strings.Join(nsubs, "/")
		root = filepath.FromSlash(rootSlash)
	}
	return root
}

//MakePath create a path from root to target.
func MakePath(to string) string {
	Root()
	return filepath.FromSlash(fmt.Sprintf("%s/%s", rootSlash, to))
}
