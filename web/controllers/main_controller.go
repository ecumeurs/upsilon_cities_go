package controllers

import (
	"net/http"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

//Home GET /
func Home(w http.ResponseWriter, req *http.Request) {
	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "index", "")
	}
}
