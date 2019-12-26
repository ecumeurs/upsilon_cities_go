package functions

import (
	"errors"
	"html/template"
	"net/http"
	"upsilon_cities_go/lib/cities/user"
	"upsilon_cities_go/lib/cities/user_log"
	"upsilon_cities_go/web/webtools"
)

//PreLoadFunctions add function at parse time.
//Note If you plan to use custom/local functions, you NEED to add them here first.
func PreLoadFunctions(t *template.Template) {
	fns := make(template.FuncMap)

	fns["IsLogged"] = func() bool { return false }
	fns["IsMap"] = func() bool { return false }
	fns["CurrentCorpID"] = func() (int, error) { return 0, errors.New("not implemented yet") }
	fns["CurrentCorpName"] = func() (string, error) { return "", errors.New("not implemented yet") }
	fns["IsAdmin"] = func() bool { return false }
	fns["CurrentUser"] = func() (*user.User, error) { return nil, errors.New("not implemented yet") }
	fns["CurrentUserID"] = func() (int, error) { return 0, errors.New("not implemented yet") }
	fns["GetRouter"] = webtools.GetRouter
	fns["CurrentCorpID"] = func() int { return 0 }
	fns["CurrentCorpName"] = func() (string, error) { return "", errors.New("not implemented yet") }

	fns["ErrorAlerts"] = func() string { return "" }
	fns["InfoAlerts"] = func() string { return "" }
	fns["WarningAlerts"] = func() string { return "" }
	fns["UserLogs"] = func() []user_log.UserLog { return make([]user_log.UserLog, 0) }

	t = t.Funcs(fns)
}

//LoadFunctions add functions to the template
//should find a way to dynamically add functions ...
func LoadFunctions(w http.ResponseWriter, req *http.Request, t *template.Template, fns template.FuncMap) {
	// add generic functions ...

	fns["IsLogged"] = IsLogged(w, req)
	fns["IsAdmin"] = IsAdmin(w, req)
	fns["CurrentCorpID"] = CurrentCorpID(w, req)
	fns["IsMap"] = IsMap(w, req)
	fns["CurrentUser"] = CurrentUser(w, req)
	fns["CurrentUserID"] = CurrentUserID(w, req)
	fns["CurrentCorpName"] = CurrentCorpName(w, req)
	fns["GetRouter"] = webtools.GetRouter
	fns["ErrorAlerts"] = ErrorAlerts(w, req)
	fns["InfoAlerts"] = InfoAlerts(w, req)
	fns["WarningAlerts"] = WarningAlerts(w, req)
	fns["UserLogs"] = UserLogs(w, req)

	t = t.Funcs(fns)
}

//IsLogged Function generator
func IsLogged(w http.ResponseWriter, req *http.Request) func() bool {
	return func() bool {
		return webtools.IsLogged(req)
	}
}

//IsMap Function generator
func IsMap(w http.ResponseWriter, req *http.Request) func() bool {
	return func() bool {
		return webtools.IsMap(req)
	}
}

//IsAdmin Function generator
func IsAdmin(w http.ResponseWriter, req *http.Request) func() bool {
	return func() bool {
		return webtools.IsAdmin(req)
	}
}

//CurrentUser Function generator
func CurrentUser(w http.ResponseWriter, req *http.Request) func() *user.User {
	return func() *user.User {
		cid, _ := webtools.CurrentUser(req)
		return cid
	}
}

//CurrentUserID Function generator
func CurrentUserID(w http.ResponseWriter, req *http.Request) func() int {
	return func() int {
		cid, _ := webtools.CurrentUserID(req)
		return cid
	}
}

//CurrentCorpID Function generator
func CurrentCorpID(w http.ResponseWriter, req *http.Request) func() int {
	return func() int {
		cid, _ := webtools.CurrentCorpID(req)
		return cid
	}
}

//CurrentCorpName Function generator
func CurrentCorpName(w http.ResponseWriter, req *http.Request) func() (string, error) {
	return func() (string, error) {
		return webtools.CurrentCorpName(req)
	}
}

//ErrorAlerts tell whether alerts marked as errors are available.
func ErrorAlerts(w http.ResponseWriter, req *http.Request) func() string {
	return func() string {
		return webtools.ErrorAlerts(req)
	}
}

//InfoAlerts tell whether alerts marked as errors are available.
func InfoAlerts(w http.ResponseWriter, req *http.Request) func() string {
	return func() string {
		return webtools.InfoAlerts(req)
	}
}

//WarningAlerts tell whether alerts marked as errors are available.
func WarningAlerts(w http.ResponseWriter, req *http.Request) func() string {
	return func() string {
		return webtools.WarningAlerts(req)
	}
}

//UserLogs fetch if available user logs.
func UserLogs(w http.ResponseWriter, req *http.Request) func() []user_log.UserLog {
	return func() []user_log.UserLog {
		res, err := webtools.UserLogs(req)
		if err != nil {
			return make([]user_log.UserLog, 0)
		}
		return res
	}
}
