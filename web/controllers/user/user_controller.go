package user_controller

import (
	"encoding/json"
	"net/http"
	"upsilon_cities_go/lib/cities/user"
	"upsilon_cities_go/lib/cities/user_log"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/web/templates"
	"upsilon_cities_go/web/webtools"
)

//Show GET /user ... Non admin version ... still must be logged ;)
func Show(w http.ResponseWriter, req *http.Request) {

	if !webtools.CheckLogged(w, req) {
		return
	}

	user, err := webtools.CurrentUser(req)

	if err != nil {
		webtools.Fail(w, req, "unable to find current user ...", "/")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(user)
	} else {
		templates.RenderTemplate(w, req, "user/show", user)
	}
}

//New GET /user/new ... Display new account screen
func New(w http.ResponseWriter, req *http.Request) {

	if webtools.IsLogged(req) {
		webtools.Fail(w, req, "must not be logged in", "/")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "user/new", "data")
	}
}

//Create POST /user ... Create new account
func Create(w http.ResponseWriter, req *http.Request) {
	if webtools.IsLogged(req) {
		webtools.Fail(w, req, "must not be logged in", "/")
		return
	}

	req.ParseForm()
	f := req.Form

	// expect a few informations:
	// login, mail, password ?

	mail := f.Get("email")
	login := f.Get("login")
	htmlpwd := f.Get("password")

	if !user.CheckLogin(login) || !user.CheckPassword(htmlpwd) || !user.CheckMail(mail) {
		webtools.Fail(w, req, "invalid parameter provided", "/users/new")
		return
	}

	usr := user.New()
	usr.Email = mail
	usr.Login = login
	pwd, err := user.HashPassword(usr.Login + htmlpwd)
	usr.Password = pwd

	if err != nil {
		webtools.Fail(w, req, "fail to hash password for database storage.", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	err = usr.Insert(dbh)
	if err != nil {
		webtools.Fail(w, req, "failed to insert user in database", "/")
		return
	}

	webtools.GetSession(req).Values["current_user_id"] = usr.ID
	webtools.GetSession(req).Values["is_admin"] = usr.Admin
	webtools.GetSession(req).Values["is_enabled"] = usr.Enabled

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.GetSession(req).AddFlash("User successfully created.", "info")
		webtools.Redirect(w, req, "/map")
	}
}

//ShowLogin GET /users/login display login screen (probably forcefully)
func ShowLogin(w http.ResponseWriter, req *http.Request) {

	if webtools.IsLogged(req) {
		webtools.Fail(w, req, "must not be logged in", "/")
		return
	}
	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "user/login", "data")
	}
}

//Login POST /users/login perform login
func Login(w http.ResponseWriter, req *http.Request) {

	if webtools.IsLogged(req) {
		webtools.Fail(w, req, "must not be logged in", "/")
		return
	}

	req.ParseForm()
	f := req.Form

	// expect a few informations:
	// login, mail, password ?

	login := f.Get("login")
	password := f.Get("password")

	dbh := db.New()
	defer dbh.Close()

	usr, err := user.ByLogin(dbh, login)

	if err != nil {
		webtools.Fail(w, req, "unable to log user in", "/")
		return
	}

	if user.CheckPasswordHash(login+password, usr.Password) {

		webtools.GetSession(req).Values["current_user_id"] = usr.ID
		webtools.GetSession(req).Values["is_admin"] = usr.Admin
		webtools.GetSession(req).Values["is_enabled"] = usr.Enabled

		usr.LogsIn(dbh)

		if usr.NeedNewPassword {
			if webtools.IsAPI(req) {
				webtools.GenerateAPIOkAndSend(w)
			} else {
				webtools.GetSession(req).AddFlash("User successfully logged in.", "info")
				webtools.Redirect(w, req, "/user/reset_password")
			}

		} else {
			if webtools.IsAPI(req) {
				webtools.GenerateAPIOkAndSend(w)
			} else {
				webtools.GetSession(req).AddFlash("User successfully logged in.", "info")
				webtools.Redirect(w, req, "/map")
			}
		}
		return
	}
	webtools.Fail(w, req, "unable to log user in", "/")
	return
}

//Logout POST /user/logout
func Logout(w http.ResponseWriter, req *http.Request) {

	if !webtools.CheckLogged(w, req) {
		return
	}
	webtools.CloseSession(req)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.Redirect(w, req, "/")
	}
}

//ShowResetPassword GET /users/reset_password ... Reset password window
func ShowResetPassword(w http.ResponseWriter, req *http.Request) {

	if !webtools.CheckLogged(w, req) {
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "user/reset_password", "")
	}
}

//ResetPassword POST /users/reset_password ... Reset password
func ResetPassword(w http.ResponseWriter, req *http.Request) {

	if !webtools.CheckLogged(w, req) {
		return
	}

	req.ParseForm()
	f := req.Form

	dbh := db.New()
	defer dbh.Close()

	usr, err := webtools.CurrentUser(req)

	if err != nil {
		webtools.Fail(w, req, "unable to find user", "/")
		return
	}

	usr.Password, err = user.HashPassword(usr.Login + f.Get("password"))
	if err != nil {
		webtools.Fail(w, req, "unable to hash user password ", "/")
		return
	}

	err = usr.UpdatePassword(dbh)
	if err != nil {
		webtools.Fail(w, req, "unable to update user password ", "/")
		return
	}

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.GetSession(req).AddFlash("User password successfully updated.", "info")
		webtools.Redirect(w, req, "/")
	}
}

//Destroy DELETE /users ... Destroy account
func Destroy(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	usr, err := webtools.CurrentUser(req)

	if err != nil {
		webtools.Fail(w, req, "unable to find user", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()
	user.Drop(dbh, usr.ID)

	webtools.CloseSession(req)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOkAndSend(w)
	} else {
		webtools.GetSession(req).AddFlash("User successfully destroyed.", "info")
		webtools.Redirect(w, req, "/")
	}
}

//Logs GET /user/logs
func Logs(w http.ResponseWriter, req *http.Request) {
	if !webtools.CheckLogged(w, req) {
		return
	}

	uid, _ := webtools.CurrentUserID(req)

	dbh := db.New()
	defer dbh.Close()

	logs := user_log.LastMessages(dbh, uid)

	if webtools.IsAPI(req) {
		webtools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(logs)
	} else {
		templates.RenderTemplate(w, req, "user/logs", logs)
	}
}
