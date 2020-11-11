package controllers

import (
	"net/http"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/webtools"
)

//Home GET /
func Home(w http.ResponseWriter, req *http.Request) {
	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "index", "")
	}
}
