package user_controller

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
		templates.RenderTemplate(w, req, "users\\index", users)
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
		templates.RenderTemplate(w, req, "users\\show", user)
	}
}

//Show GET /user ... Non admin version ... still must be logged ;)
func Show(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
		return
	}

	user, err := tools.CurrentUser(req)

	if err != nil {
		tools.Fail(w, req, "unable to find current user ...", "/")
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOk(w)
		json.NewEncoder(w).Encode(user)
	} else {
		templates.RenderTemplate(w, req, "users\\show", user)
	}
}

//New GET /user/new ... Display new account screen
func New(w http.ResponseWriter, req *http.Request) {

	if tools.IsLogged(req) {
		tools.Fail(w, req, "must not be logged in", "/")
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "users\\new", "data")
	}
}

//Create POST /user ... Create new account
func Create(w http.ResponseWriter, req *http.Request) {

	if tools.IsLogged(req) {
		tools.Fail(w, req, "must not be logged in", "/")
	}

	req.ParseForm()
	f := req.Form

	// expect a few informations:
	// login, mail, password ?

	usr := user.New()
	usr.Email = f.Get("email")
	usr.Login = f.Get("login")

	pwd, err := user.HashPassword(usr.Login + f.Get("password"))
	usr.Password = pwd

	if err != nil {
		tools.Fail(w, req, "fail to hash password for database storage.", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()

	err = usr.Insert(dbh)
	if err != nil {
		tools.Fail(w, req, "failed to insert user in database", "/")
		return
	}

	tools.GetSession(req).Values["current_user_id"] = usr.ID
	tools.GetSession(req).Values["is_admin"] = usr.Admin
	tools.GetSession(req).Values["is_enabled"] = usr.Enabled

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.GetSession(req).AddFlash("User successfully created.")
		tools.Redirect(w, req, "/map")
	}
}

//ShowLogin GET /users/login display login screen (probably forcefully)
func ShowLogin(w http.ResponseWriter, req *http.Request) {
	if tools.IsLogged(req) {
		tools.Fail(w, req, "must not be logged in", "/")
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "users\\login", "data")
	}
}

//Login POST /users/login perform login
func Login(w http.ResponseWriter, req *http.Request) {

	if tools.IsLogged(req) {
		tools.Fail(w, req, "must not be logged in", "/")
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
		tools.Fail(w, req, "unable to log user in", "/")
		return
	}

	if user.CheckPasswordHash(login+password, usr.Password) {

		tools.GetSession(req).Values["current_user_id"] = usr.ID
		tools.GetSession(req).Values["is_admin"] = usr.Admin
		tools.GetSession(req).Values["is_enabled"] = usr.Enabled

		usr.LogsIn(dbh)

		if usr.NeedNewPassword {
			if tools.IsAPI(req) {
				tools.GenerateAPIOkAndSend(w)
			} else {
				tools.GetSession(req).AddFlash("User successfully logged in.")
				tools.Redirect(w, req, "/users/reset_password")
			}

		} else {
			if tools.IsAPI(req) {
				tools.GenerateAPIOkAndSend(w)
			} else {
				tools.GetSession(req).AddFlash("User successfully logged in.")
				tools.Redirect(w, req, "/map")
			}
		}
	}
	tools.Fail(w, req, "unable to log user in", "/")
	return
}

//Logout POST /user/logout
func Logout(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}
	tools.CloseSession(req)

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.Redirect(w, req, "/")
	}
}

//ShowResetPassword GET /users/reset_password ... Reset password window
func ShowResetPassword(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		templates.RenderTemplate(w, req, "users\\reset_password", "")
	}
}

//ResetPassword POST /users/reset_password ... Reset password
func ResetPassword(w http.ResponseWriter, req *http.Request) {

	if !tools.IsLogged(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}

	req.ParseForm()
	f := req.Form

	dbh := db.New()
	defer dbh.Close()

	usr, err := tools.CurrentUser(req)

	if err != nil {
		tools.Fail(w, req, "unable to find user", "/")
		return
	}

	usr.Password, err = user.HashPassword(usr.Login + f.Get("password"))
	if err != nil {
		tools.Fail(w, req, "unable to hash user password ", "/")
		return
	}

	err = usr.UpdatePassword(dbh)
	if err != nil {
		tools.Fail(w, req, "unable to update user password ", "/")
		return
	}

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.GetSession(req).AddFlash("User password successfully updated.")
		tools.Redirect(w, req, "/")
	}
}

//Destroy DELETE /users ... Destroy account
func Destroy(w http.ResponseWriter, req *http.Request) {
	if !tools.IsLogged(req) || !tools.IsAdmin(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}

	usr, err := tools.CurrentUser(req)

	if err != nil {
		tools.Fail(w, req, "unable to find user", "/")
		return
	}

	dbh := db.New()
	defer dbh.Close()
	user.Drop(dbh, usr.ID)

	tools.CloseSession(req)

	if tools.IsAPI(req) {
		tools.GenerateAPIOkAndSend(w)
	} else {
		tools.GetSession(req).AddFlash("User successfully destroyed.")
		tools.Redirect(w, req, "/")
	}
}

//AdminReset POST /admin/users/:user_id/reset
func AdminReset(w http.ResponseWriter, req *http.Request) {
	if !tools.IsAdmin(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}
	id, err := tools.GetInt(req, "user_id")

	if err != nil {
		tools.Fail(w, req, "unable to parse user_id", "/")
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
		tools.GetSession(req).AddFlash("User successfully destroyed.")
		tools.Redirect(w, req, "/")
	}
}

//AdminDestroy DELETE /admin/users/:user_id ... Destroy account
func AdminDestroy(w http.ResponseWriter, req *http.Request) {
	if !tools.IsAdmin(req) {
		tools.Fail(w, req, "must be logged in", "/")
	}
	id, err := tools.GetInt(req, "user_id")

	if err != nil {
		tools.Fail(w, req, "unable to parse user_id", "/")
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
		tools.GetSession(req).AddFlash("User successfully destroyed.")
		tools.Redirect(w, req, "/")
	}
}