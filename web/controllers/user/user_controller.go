package user_controller

import (
	"encoding/json"
	"net/http"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

//Index GET /admin/users ... Must be logged as admin to get here ;)
func Index(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		templates.RenderTemplate(w, req, "users\\index", "data")
	}
}

//AdminShow GET /admin/users/:user_id ... Must be logged as admin to get here ;)
func AdminShow(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		templates.RenderTemplate(w, req, "users\\show", "data")
	}
}

//Show GET /users/:user_id ... Non admin version ... still must be logged ;)
func Show(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		templates.RenderTemplate(w, req, "users\\show", "data")
	}
}

//New GET /users/new ... Display new account screen
func New(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		templates.RenderTemplate(w, req, "users\\new", "data")
	}
}

//Create POST /users ... Create new account
func Create(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		tools.GetSession(w, req).AddFlash("User successfully created.")
		tools.Redirect(w, req, "/map")
	}
}

//ShowLogin GET /users/login display login screen (probably forcefully)
func ShowLogin(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		tools.GetSession(w, req).AddFlash("User successfully logged in.")
		templates.RenderTemplate(w, req, "users\\login", "data")
	}
}

//Login POST /users/login perform login
func Login(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		tools.GetSession(w, req).AddFlash("User successfully logged in.")
		tools.Redirect(w, req, "/map")
	}
}

//ShowResetPassword GET /users/:user_id/reset_password ... Reset password window
func ShowResetPassword(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		templates.RenderTemplate(w, req, "users\\reset_password", "data")
	}
}

//ResetPassword POST /users/:user_id/reset_password ... Reset password
func ResetPassword(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		templates.RenderTemplate(w, req, "users\\index", "data")
	}
}

//Destroy DELETE /users:user_id ... Destroy account, either as admin or logged in ;)
func Destroy(w http.ResponseWriter, req *http.Request) {

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode("data")
	} else {
		tools.GetSession(w, req).AddFlash("User successfully destroyed.")
		tools.Redirect(w, req, "/")
	}
}
