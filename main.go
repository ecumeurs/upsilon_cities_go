package main

import (
	"log"
	"math/rand"
	"time"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/generator"
	"upsilon_cities_go/web"
)

func main() {
	rand.Seed(time.Now().Unix())
	generator.Init()
	// ensure that in memory storage is fine.
	city_manager.InitManager()
	grid_manager.InitManager()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	handler := db.New()
	db.CheckVersion(handler)

	r := web.RouterSetup()
	web.ListenAndServe(r)

	defer handler.Close()

}
