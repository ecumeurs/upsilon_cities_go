package main

import (
	"log"
	"math/rand"
	"time"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web"
)

func main() {
	rand.Seed(time.Now().Unix())

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	handler := db.New()
	db.CheckVersion(handler)

	r := web.RouterSetup()
	web.ListenAndServe(r)

	defer handler.Close()

}
