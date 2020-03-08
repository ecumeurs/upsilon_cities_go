package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city/producer_generator"
	"upsilon_cities_go/lib/cities/city/resource_generator"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/map/grid_manager"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/gameplay"
	"upsilon_cities_go/lib/misc/config/system"
	"upsilon_cities_go/lib/misc/generator"
	"upsilon_cities_go/web"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/webtools"
)

func main() {
	rand.Seed(time.Now().Unix())

	shouldLogInFile := flag.Bool("log", false, "moves logs to logs.txt file.")
	flag.Parse()
	if *shouldLogInFile {
		f, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		log.SetOutput(f)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	system.LoadConf()
	gameplay.LoadConf()

	tools.InitCycle()
	// ensure that in memory storage is fine.
	city_manager.InitManager()
	grid_manager.InitManager()
	caravan_manager.InitManager()
	corporation_manager.InitManager()

	generator.CreateSampleFile()
	generator.Load()

	producer_generator.CreateSampleFile()
	producer_generator.Load()

	resource_generator.Load()
	caravan.Init()
	handler := db.New()
	db.CheckVersion(handler)
	handler.Close()

	router := web.RouterSetup()
	webtools.SetRouter(router)
	templates.LoadTemplates()
	web.ListenAndServe(router)

}
