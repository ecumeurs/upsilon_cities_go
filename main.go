package main

import (
	"math/rand"
	"time"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web"
)

func main() {
	rand.Seed(time.Now().Unix())
	// grid := grid.New()
	// log.Printf("Resulting grid: ")
	// log.Printf("\n%v\n", grid)
	// return

	handler := db.New()
	db.CheckVersion(handler)

	// testDB(handler)
	r := web.RouterSetup()
	web.ListenAndServe(r)

	defer handler.Close()

}
