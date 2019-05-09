package main

import (
	"log"
	"math/rand"
	"time"
	"upsilon_cities_go/lib/cities/city/producer"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/grid_manager"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/generator"
	"upsilon_cities_go/web"
)

func main() {
	rand.Seed(time.Now().Unix())

	tools.InitCycle()
	// ensure that in memory storage is fine.

	city_manager.InitManager()
	grid_manager.InitManager()

	generator.CreateSampleFile()
	generator.Init()

	producer.CreateSampleFile()
	producer.Load()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	handler := db.New()
	db.CheckVersion(handler)

	r := web.RouterSetup()
	web.ListenAndServe(r)

	defer handler.Close()

}
