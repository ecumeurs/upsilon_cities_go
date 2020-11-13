package admin_controller

import (
	"encoding/json"
	"net/http"
	"os"
	"upsilon_cities_go/lib/cities/user"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/webtools"
)

//Index GET /admin/users ... Must be logged as admin to get here ;)
func Index(w http.ResponseWriter, req *http.Request) {
	if !webtools.IsAdmin(req) {
		webtools.Redirect(w, req, "/")
		return
	}
	dbh := db.New()
	defer dbh.Close()
	users := user.All(dbh)
	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(users)
	} else {
		templates.RenderTemplate(w, req, "admin/index", users)
	}
}

//AdminTools GET /admin/users ... Must be logged as admin to get here ;)
func AdminTools(w http.ResponseWriter, req *http.Request) {
	if !webtools.IsAdmin(req) {
		webtools.Redirect(w, req, "/")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
	} else {
		templates.RenderTemplate(w, req, "admin/tools", "")
	}
}

//ReloadDb GET /admin/users ... Must be logged as admin to get here ;)
func ReloadDb(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckAdmin(w, req) {
		return
	}

	dbh := db.New()
	defer dbh.Close()

	db.FlushDatabase(dbh)

	os.Exit(5001)

}

//ReloadServer GET /admin/users ... Must be logged as admin to get here ;)
func ReloadServer(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckAdmin(w, req) {
		return
	}

	os.Exit(5002)
}

//AdminShow GET /admin/users/:user_id ... Must be logged as admin to get here ;)
func AdminShow(w http.ResponseWriter, req *http.Request) {
	if !webtools.IsAdmin(req) {
		webtools.Redirect(w, req, "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()
	id, err := webtools.GetInt(req, "user_id")

	if err != nil {
		webtools.Fail(w, req, "invalid parameter provided", "/admin/users/")
		return
	}

	user, err := user.ByID(dbh, id)

	if err != nil {
		webtools.Fail(w, req, "fail to find requested user", "/admin/users/")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(user)
	} else {
		templates.RenderTemplate(w, req, "user/show", user)
	}
}

//AdminReset POST /admin/users/:user_id/reset
func AdminReset(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckAdmin(w, req) {
		return
	}

	id, err := webtools.GetInt(req, "user_id")

	if err != nil {
		webtools.Fail(w, req, "unable to parse user_id", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	usr, err := user.ByID(dbh, id)

	if err != nil {
		webtools.Fail(w, req, "unable to find user", "/")
		return
	}

	usr.NeedNewPassword = true
	usr.Update(dbh)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.GetSession(req).AddFlash("User successfully destroyed.", "info")
		webtools.Redirect(w, req, "/")
	}
}

//AdminReset POST /admin/users/:user_id/reset
func Lock(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckAdmin(w, req) {
		return
	}

	id, err := webtools.GetInt(req, "user_id")
	state, err := webtools.GetInt(req, "user_state")

	if err != nil {
		webtools.Fail(w, req, "unable to parse user_id", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	usr, err := user.ByID(dbh, id)

	if err != nil {
		webtools.Fail(w, req, "unable to find user", "/")
		return
	}

	usr.Enabled = state == 1
	usr.Update(dbh)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.GetSession(req).AddFlash("User successfully destroyed.", "info")
		webtools.Redirect(w, req, "/admin/users")
	}
}

//AdminDestroy DELETE /admin/users/:user_id ... Destroy account
func AdminDestroy(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckAdmin(w, req) {
		return
	}

	id, err := webtools.GetInt(req, "user_id")

	if err != nil {
		webtools.Fail(w, req, "unable to parse user_id", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	usr, err := user.ByID(dbh, id)

	if err != nil {
		webtools.Fail(w, req, "unable to find user", "/")
		return
	}

	user.Drop(dbh, usr.ID)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.GetSession(req).AddFlash("User successfully destroyed.", "info")
		webtools.Redirect(w, req, "/")
	}
}
