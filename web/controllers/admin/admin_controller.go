package admin_controller

import (
	"encoding/json"
	"net/http"
	"upsilon_cities_go/lib/cities/user"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/tools"
)

//Index GET /admin/users ... Must be logged as admin to get here ;)
func Index(w http.ResponseWriter, req *http.Request) {
	if !tools.IsAdmin(req) {
		tools.Redirect(w, req, "/")
		return
	}
	dbh := db.New()
	defer dbh.Close()
	users := user.All(dbh)
	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(users)
	} else {
		templates.RenderTemplate(w, req, "admin/index", users)
	}
}

//AdminShow GET /admin/users/:user_id ... Must be logged as admin to get here ;)
func AdminShow(w http.ResponseWriter, req *http.Request) {
	if !tools.IsAdmin(req) {
		tools.Redirect(w, req, "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()
	id, err := tools.GetInt(req, "user_id")

	if err != nil {
		tools.Fail(w, req, "invalid parameter provided", "/admin/users/")
		return
	}

	user, err := user.ByID(dbh, id)

	if err != nil {
		tools.Fail(w, req, "fail to find requested user", "/admin/users/")
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(user)
	} else {
		templates.RenderTemplate(w, req, "user/show", user)
	}
}

//AdminReset POST /admin/users/:user_id/reset
func AdminReset(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckAdmin(w, req) {
		return
	}

	id, err := tools.GetInt(req, "user_id")

	if err != nil {
		tools.Fail(w, req, "unable to parse user_id", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	usr, err := user.ByID(dbh, id)

	if err != nil {
		tools.Fail(w, req, "unable to find user", "/")
		return
	}

	usr.NeedNewPassword = true
	usr.Update(dbh)

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.GetSession(req).AddFlash("User successfully destroyed.", "info")
		tools.Redirect(w, req, "/")
	}
}

//AdminReset POST /admin/users/:user_id/reset
func Lock(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckAdmin(w, req) {
		return
	}

	id, err := tools.GetInt(req, "user_id")
	state, err := tools.GetInt(req, "user_state")

	if err != nil {
		tools.Fail(w, req, "unable to parse user_id", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	usr, err := user.ByID(dbh, id)

	if err != nil {
		tools.Fail(w, req, "unable to find user", "/")
		return
	}

	usr.Enabled = state == 1
	usr.Update(dbh)

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.GetSession(req).AddFlash("User successfully destroyed.", "info")
		tools.Redirect(w, req, "/admin/users")
	}
}

//AdminDestroy DELETE /admin/users/:user_id ... Destroy account
func AdminDestroy(w http.ResponseWriter, req *http.Request) {
	if !tools.CheckAdmin(w, req) {
		return
	}

	id, err := tools.GetInt(req, "user_id")

	if err != nil {
		tools.Fail(w, req, "unable to parse user_id", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	usr, err := user.ByID(dbh, id)

	if err != nil {
		tools.Fail(w, req, "unable to find user", "/")
		return
	}

	user.Drop(dbh, usr.ID)

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.GetSession(req).AddFlash("User successfully destroyed.", "info")
		tools.Redirect(w, req, "/")
	}
}