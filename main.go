package main

import (
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web"
)

func main() {
	handler := db.New()
	// testDB(handler)
	r := web.RouterSetup()
	web.ListenAndServe(r)

	defer handler.Close()

}
