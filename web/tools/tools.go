package tools

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/tools"
	"upsilon_cities_go/lib/cities/user"
	"upsilon_cities_go/lib/cities/user_log"
	"upsilon_cities_go/lib/db"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	sessions "github.com/gorilla/sessions"
)

var defaultRouter *mux.Router

//GetRouter router.
func GetRouter() *mux.Router {
	return defaultRouter
}

//SetRouter router.
func SetRouter(m *mux.Router) {
	defaultRouter = m
}

//GetSession from store
func GetSession(r *http.Request) (session *sessions.Session) {
	return context.Get(r, "session").(*sessions.Session)
}

//CloseSession  ;)
func CloseSession(r *http.Request) {
	GetSession(r).Options.MaxAge = -1
}

// IsAPI Tell whether request requires API reply or not.
func IsAPI(req *http.Request) bool {
	return strings.Contains(req.URL.String(), "/api/")
}

// IsMap Tell whether request open a map.
func IsMap(req *http.Request) bool {
	return strings.Contains(req.URL.String(), "/map/")
}

//CurrentUser fetch current user.
func CurrentUser(req *http.Request) (*user.User, error) {
	if IsLogged(req) {
		dbh := db.New()
		defer dbh.Close()
		us, err := user.ByID(dbh, GetSession(req).Values["current_user_id"].(int))
		if err != nil {
			return nil, err
		}
		return us, nil
	}
	return nil, errors.New("no user logged in")
}

//CurrentUserID fetch current user.
func CurrentUserID(req *http.Request) (int, error) {
	if IsLogged(req) {
		return GetSession(req).Values["current_user_id"].(int), nil
	}
	return 0, errors.New("no user logged in")
}

//UserLogs fetch if available user logs.
func UserLogs(req *http.Request) ([]user_log.UserLog, error) {
	if !IsLogged(req) {
		return nil, errors.New("not logged so no logs")
	}
	dbh := db.New()
	defer dbh.Close()
	uid, _ := CurrentUserID(req)
	return user_log.Since(dbh, uid, tools.AboutNow(-300)), nil
}

//IsLogged tell whether user is logged or not.
func IsLogged(req *http.Request) bool {
	_, found := GetSession(req).Values["current_user_id"]
	return found
}

//IsAdmin tell whether user is logged or not.
func IsAdmin(req *http.Request) bool {
	_, found := GetSession(req).Values["is_admin"]
	return found
}

//CurrentCorpID tell whether user is logged or not.
func CurrentCorpID(req *http.Request) (int, error) {
	corp, found := GetSession(req).Values["current_corp_id"]
	if !found {
		return 0, errors.New("not found")
	}
	return corp.(int), nil
}

//CurrentCorpName tell whether user is logged or not.
func CurrentCorpName(req *http.Request) (string, error) {
	crp, err := CurrentCorp(req)
	if err != nil {
		return "", errors.New("not found")
	}
	return crp.Get().Name, nil
}

//CurrentCorp fetch current user.
func CurrentCorp(req *http.Request) (*corporation_manager.Handler, error) {
	if IsLogged(req) {
		dbh := db.New()
		defer dbh.Close()
		corpID, err := CurrentCorpID(req)
		if err != nil {
			return nil, err
		}

		return corporation_manager.GetCorporationHandler(corpID)
	}
	return nil, errors.New("no user logged in")
}

// GetInt parse request to get int value.
func GetInt(req *http.Request, key string) (int, error) {
	vars := mux.Vars(req)
	value, err := strconv.Atoi(vars[key])
	if err != nil {
		log.Printf("Web: requested key: %s , not found in: %s", key, req.URL)
		return 0, errors.New("Invalid key requested")
	}
	return value, nil
}

// GetIntSilent parse request to get int value.
func GetIntSilent(req *http.Request, key string) (int, error) {
	vars := mux.Vars(req)
	value, err := strconv.Atoi(vars[key])
	if err != nil {
		return 0, errors.New("Invalid key requested")
	}
	return value, nil
}

// GetString parse request to get int value.
func GetString(req *http.Request, key string) (value string, found bool) {
	vars := mux.Vars(req)
	value, found = vars[key]
	return
}

//Fail fails current request with API and redirect with Web; should set session error ;) but can't right now ...
func Fail(w http.ResponseWriter, req *http.Request, err string, backRoute string) {
	log.Printf("Web: Failed access to %s due to %s", req.URL.String(), err)
	if IsAPI(req) {
		GenerateAPIError(w, err)
	} else {
		GetSession(req).AddFlash(err, "error")
		Redirect(w, req, backRoute)
	}
}

//Redirect user to targeted page. If route is empty, will redirect to referer. (calling webpage)
func Redirect(w http.ResponseWriter, req *http.Request, route string) {
	log.Printf("Web: Redirecting to %s", route)

	if route == "" {
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	} else {
		http.Redirect(w, req, route, http.StatusSeeOther)
	}
}

//ErrorAlerts returns error alerts
func ErrorAlerts(req *http.Request) string {
	res := make([]string, 0)
	for _, v := range GetSession(req).Flashes("error") {
		res = append(res, v.(string))
	}

	return strings.Join(res, ",")
}

//InfoAlerts returns error alerts
func InfoAlerts(req *http.Request) string {
	res := make([]string, 0)
	for _, v := range GetSession(req).Flashes("info") {
		res = append(res, v.(string))
	}

	return strings.Join(res, ",")

}

//WarningAlerts returns error alerts
func WarningAlerts(req *http.Request) string {
	res := make([]string, 0)
	for _, v := range GetSession(req).Flashes("warning") {
		res = append(res, v.(string))
	}

	return strings.Join(res, ",")
}

// HasValue tell whether value is present or not.
func HasValue(req *http.Request, key string) bool {
	vars := mux.Vars(req)
	_, ok := vars[key]
	return ok
}

//CheckLogged if not logged return false and fails request
func CheckLogged(w http.ResponseWriter, req *http.Request) bool {
	if !IsLogged(req) {
		Fail(w, req, "must be logged to access this content.", "")
		return false
	}
	return true
}

//CheckAPI if not logged return false and fails request
func CheckAPI(w http.ResponseWriter, req *http.Request) bool {
	if !IsAPI(req) {
		Fail(w, req, "content is only accessible in API", "")
		return false
	}
	return true
}

//CheckWeb if not logged return false and fails request
func CheckWeb(w http.ResponseWriter, req *http.Request) bool {
	if IsAPI(req) {
		Fail(w, req, "content is only accessible in WEB", "")
		return false
	}
	return true
}

//CheckAdmin if not logged return false and fails request
func CheckAdmin(w http.ResponseWriter, req *http.Request) bool {
	if !IsAdmin(req) {
		Fail(w, req, "content is for admin eyes only", "")
		return false
	}
	return true
}

// GenerateAPIError generate a simple JSON reply with error message provided.
func GenerateAPIError(w http.ResponseWriter, message string) {
	var repm = make(map[string]string)
	repm["status"] = "error"
	repm["error"] = message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(repm)
}

// GenerateAPIOkAndSend generate a simple JSON reply with status: ok.
func GenerateAPIOkAndSend(w http.ResponseWriter) {
	var repm = make(map[string]string)
	repm["status"] = "ok"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(repm)
}

// GenerateAPIOk generate a simple JSON reply with status: ok.
func GenerateAPIOk(w http.ResponseWriter) map[string]string {
	var repm = make(map[string]string)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	repm["status"] = "ok"
	return repm
}
