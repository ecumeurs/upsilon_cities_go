package web

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"upsilon_cities_go/lib/cities/caravan"
	"upsilon_cities_go/lib/cities/caravan_manager"
	"upsilon_cities_go/lib/cities/city"
	"upsilon_cities_go/lib/cities/city_manager"
	"upsilon_cities_go/lib/cities/corporation"
	"upsilon_cities_go/lib/cities/corporation_manager"
	"upsilon_cities_go/lib/cities/map/grid_manager"
	"upsilon_cities_go/lib/db"
	"upsilon_cities_go/lib/misc/config/system"
	controllers "upsilon_cities_go/web/controllers"
	admin_controller "upsilon_cities_go/web/controllers/admin"
	crv_controller "upsilon_cities_go/web/controllers/caravan"
	city_controller "upsilon_cities_go/web/controllers/city"
	corp_controller "upsilon_cities_go/web/controllers/corporation"
	grid_controller "upsilon_cities_go/web/controllers/grid"
	user_controller "upsilon_cities_go/web/controllers/user"
	"upsilon_cities_go/web/webtools"

	"github.com/antonlindstrom/pgstore"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

var store *pgstore.PGStore

// RouterSetup Prepare routing.
func RouterSetup() *mux.Router {
	r := mux.NewRouter()

	dbh := db.New()
	var err error
	store, err = pgstore.NewPGStoreFromPool(dbh.Raw(), []byte(system.Get("http_session_secret_key", "12345678912345678912345678912345")))
	if err != nil {
		// failed to find a store in there ...
		log.Fatalf("Session: Failed to initialize session for web request ... %s", err)
	}

	// Run a background goroutine to clean up expired sessions from the database.
	// defer store.StopCleanup(store.Cleanup(time.Minute * 5))

	// ensure session knows how to keep some complex data types.
	initConverters()

	sessionned := r.PathPrefix("").Subrouter()

	sessionned.HandleFunc("", controllers.Home).Methods("GET")
	sessionned.HandleFunc("/", controllers.Home).Methods("GET")

	// CRUD /maps
	sessionned.HandleFunc("/map", grid_controller.Index).Methods("GET")
	sessionned.HandleFunc("/map", grid_controller.Create).Methods("POST")

	maps := sessionned.PathPrefix("/map/{map_id}").Subrouter()
	maps.HandleFunc("", grid_controller.Show).Methods("GET")
	maps.HandleFunc("/select_corporation", grid_controller.ShowSelectableCorporation).Methods("GET")
	maps.HandleFunc("/select_corporation", grid_controller.SelectCorporation).Methods("POST")
	maps.HandleFunc("/cities", city_controller.Index).Methods("GET")

	// ensure map get generated ...
	maps.Use(mapMw)

	city := sessionned.PathPrefix("/city/{city_id}").Subrouter()
	city.HandleFunc("", city_controller.Show).Methods("GET")
	city.HandleFunc("/give/{item}", city_controller.Give).Methods("POST")
	city.HandleFunc("/drop/{item}", city_controller.Drop).Methods("POST")
	city.HandleFunc("/sell/{item}", city_controller.Sell).Methods("POST")
	city.HandleFunc("/producer/{producer_id}/{action}", city_controller.ProducerUpgrade).Methods("POST")

	// ensure map get generated ...
	city.Use(mapMw)

	corporation := sessionned.PathPrefix("/corporation/{corp_id}").Subrouter()
	corporation.HandleFunc("", corp_controller.Show).Methods("GET")

	// ensure map get generated ...
	corporation.Use(mapMw)

	caravan := sessionned.PathPrefix("/caravan").Subrouter()
	// caravan related stuff
	caravan.HandleFunc("", crv_controller.Index).Methods("GET")
	caravan.HandleFunc("/new/{city_id}", crv_controller.New).Methods("GET")
	caravan.HandleFunc("/{crv_id}", crv_controller.Show).Methods("GET")
	caravan.HandleFunc("/{crv_id}/accept", crv_controller.Accept).Methods("POST")
	caravan.HandleFunc("/{crv_id}/reject", crv_controller.Reject).Methods("POST")
	caravan.HandleFunc("/{crv_id}/abort", crv_controller.Abort).Methods("POST")
	caravan.HandleFunc("/{crv_id}/counter", crv_controller.GetCounter).Methods("POST")
	caravan.HandleFunc("/{crv_id}/counter", crv_controller.PostCounter).Methods("POST")
	caravan.HandleFunc("/{crv_id}/drop", crv_controller.Drop).Methods("POST")

	// ensure map get generated ...
	caravan.Use(mapMw)

	// Interface Admin
	usr := sessionned.PathPrefix("/user").Subrouter()
	usr.HandleFunc("", user_controller.Show).Methods("GET")
	usr.HandleFunc("/new", user_controller.New).Methods("GET")
	usr.HandleFunc("", user_controller.Create).Methods("POST")
	usr.HandleFunc("/login", user_controller.ShowLogin).Methods("GET")
	usr.HandleFunc("/login", user_controller.Login).Methods("POST")
	usr.HandleFunc("/logs", user_controller.Logs).Methods("GET")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("GET")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("POST")
	usr.HandleFunc("/reset_password", user_controller.ShowResetPassword).Methods("GET")
	usr.HandleFunc("/reset_password", user_controller.ResetPassword).Methods("POST")
	usr.HandleFunc("", user_controller.Destroy).Methods("DELETE")

	// Interface Admin
	admin := sessionned.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("", admin_controller.Index).Methods("GET")
	admin.HandleFunc("/tools", admin_controller.AdminTools).Methods("GET")
	admin.HandleFunc("/tools/rdb", admin_controller.ReloadDb).Methods("DELETE")
	admin.HandleFunc("/tools/rsrv", admin_controller.ReloadServer).Methods("DELETE")
	admin.HandleFunc("/users/{user_id}", admin_controller.AdminShow).Methods("GET")
	admin.HandleFunc("/users/{user_id}", admin_controller.AdminShow).Methods("GET")
	admin.HandleFunc("/users/{user_id}/reset", admin_controller.AdminReset).Methods("POST")
	admin.HandleFunc("/users/{user_id}", admin_controller.AdminDestroy).Methods("DELETE")
	admin.HandleFunc("/users/{user_id}/state/{user_state}", admin_controller.Lock).Methods("POST")

	// JSON Access ...
	jsonAPI := sessionned.PathPrefix("/api").Subrouter()
	jsonAPI.HandleFunc("/map", grid_controller.Index).Methods("GET")
	jsonAPI.HandleFunc("/map", grid_controller.Create).Methods("POST")

	maps = jsonAPI.PathPrefix("/map/{map_id}").Subrouter()
	maps.HandleFunc("", grid_controller.Show).Methods("GET")
	maps.HandleFunc("", grid_controller.Destroy).Methods("DELETE")
	maps.HandleFunc("/select_corporation", grid_controller.ShowSelectableCorporation).Methods("GET")
	maps.HandleFunc("/select_corporation", grid_controller.SelectCorporation).Methods("POST")
	maps.HandleFunc("/cities", city_controller.Index).Methods("GET")
	maps.HandleFunc("/city/X/{x_loc}/Y/{y_loc}", city_controller.IDShow).Methods("GET")

	// ensure map get generated ...
	maps.Use(mapMw)

	city = jsonAPI.PathPrefix("/city/{city_id}").Subrouter()
	city.HandleFunc("", city_controller.Show).Methods("GET")
	city.HandleFunc("/give/{item}", city_controller.Give).Methods("POST")
	city.HandleFunc("/drop/{item}", city_controller.Drop).Methods("POST")
	city.HandleFunc("/sell/{item}", city_controller.Sell).Methods("POST")
	city.HandleFunc("/producer/{producer_id}/{action}/{product}", city_controller.ProducerUpgrade).Methods("POST")

	// ensure map get generated ...
	city.Use(mapMw)

	usr = jsonAPI.PathPrefix("/user").Subrouter()
	usr.HandleFunc("", user_controller.Show).Methods("GET")
	usr.HandleFunc("/new", user_controller.New).Methods("GET")
	usr.HandleFunc("", user_controller.Create).Methods("POST")
	usr.HandleFunc("/login", user_controller.ShowLogin).Methods("GET")
	usr.HandleFunc("/logs", user_controller.Logs).Methods("GET")
	usr.HandleFunc("/login", user_controller.Login).Methods("POST")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("GET")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("POST")
	usr.HandleFunc("/reset_password", user_controller.ShowResetPassword).Methods("GET")
	usr.HandleFunc("/reset_password", user_controller.ResetPassword).Methods("POST")
	usr.HandleFunc("", user_controller.Destroy).Methods("DELETE")

	corporation = jsonAPI.PathPrefix("/corporation/{corp_id}").Subrouter()
	corporation.HandleFunc("/", corp_controller.Show).Methods("GET")

	// ensure map get generated ...
	corporation.Use(mapMw)

	caravan = jsonAPI.PathPrefix("/caravan").Subrouter()
	// caravan related stuff
	caravan.HandleFunc("", crv_controller.Index).Methods("GET")
	caravan.HandleFunc("/new/{city_id}", crv_controller.New).Methods("GET")
	caravan.HandleFunc("", crv_controller.Create).Methods("POST")
	caravan.HandleFunc("/{crv_id}", crv_controller.Show).Methods("GET")
	caravan.HandleFunc("/{crv_id}/accept", crv_controller.Accept).Methods("POST")
	caravan.HandleFunc("/{crv_id}/reject", crv_controller.Reject).Methods("POST")
	caravan.HandleFunc("/{crv_id}/abort", crv_controller.Abort).Methods("POST")
	caravan.HandleFunc("/{crv_id}/counter", crv_controller.GetCounter).Methods("POST")
	caravan.HandleFunc("/{crv_id}/counter", crv_controller.PostCounter).Methods("POST")
	caravan.HandleFunc("/{crv_id}/drop", crv_controller.Drop).Methods("POST")

	// ensure map get generated ...
	caravan.Use(mapMw)

	admin = jsonAPI.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("", admin_controller.Index).Methods("GET")
	admin.HandleFunc("/tools/rdb", admin_controller.ReloadDb).Methods("DELETE")
	admin.HandleFunc("/tools/rsrv", admin_controller.ReloadServer).Methods("DELETE")
	admin.HandleFunc("/users/{user_id}", admin_controller.AdminShow).Methods("GET")
	admin.HandleFunc("/users/{user_id}/reset", admin_controller.AdminReset).Methods("POST")
	admin.HandleFunc("/users/{user_id}", admin_controller.AdminDestroy).Methods("DELETE")
	admin.HandleFunc("/users/{user_id}/state/{user_state}", admin_controller.Lock).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(system.MakePath(system.Get("web_static_files", "web/static"))))))

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.FromSlash(fmt.Sprintf("%s/img/favicon.ico", system.MakePath(system.Get("web_static_files", "web/static")))))
	})

	r.Use(logResultMw)
	r.Use(loggingMw)
	sessionned.Use(sessionMw)

	return r
}

// initialize "gob" that handle struct serialization for session.
// see: https://www.gorillatoolkit.org/pkg/sessions#overview
func initConverters() {

}

// mapMw ensure map is loaded.
func mapMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mid, err := webtools.GetIntSilent(r, "map_id")
		if err == nil {
			grid_manager.GetGridHandler(mid)
		}
		cid, err := webtools.GetIntSilent(r, "city_id")
		if err == nil {
			_, err := city_manager.GetCityHandler(cid)
			if err != nil {
				dbh := db.New()
				defer dbh.Close()
				c, err := city.ByID(dbh, cid)
				if err != nil {
					// targets illegal city ...
				} else {
					grid_manager.GetGridHandler(c.MapID)
				}
			}
		}
		crvID, err := webtools.GetIntSilent(r, "crv_id")
		if err == nil {
			_, err := caravan_manager.GetCaravanHandler(crvID)
			if err != nil {
				dbh := db.New()
				defer dbh.Close()
				crv, err := caravan.ByID(dbh, crvID)
				if err != nil {
					// targets illegal caravan ...
				} else {
					grid_manager.GetGridHandler(crv.MapID)
				}
			}
		}
		corpID, err := webtools.GetIntSilent(r, "corp_id")
		if err == nil {
			_, err := corporation_manager.GetCorporationHandler(corpID)
			if err != nil {
				dbh := db.New()
				defer dbh.Close()
				corp, err := corporation.ByID(dbh, corpID)
				if err != nil {
					// targets illegal caravan ...
				} else {
					grid_manager.GetGridHandler(corp.MapID)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

//sessionMw start a session
func sessionMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get a session.
		session, err := store.Get(r, "session-key")

		if err != nil {
			log.Fatalf(err.Error())
		}

		context.Set(r, "session", session)

		next.ServeHTTP(w, r)

		// Must save session before replying.
		// session := GetSession(req)
		// log.Printf("saving session: content %v", session.Values)
		// if err := session.Save(req, w); err != nil {
		//	log.Printf("Error saving session: content %v", session.Values)
		//
		//	log.Fatalf("Error saving session: %v", err)
		//}
	})
}

// loggingMw tell what route has been called.
func loggingMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Web: Received request: %s %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

// loggingMw tell what route has been called.
func logResultMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, req)
		log.Printf(
			"Web: %s %s (code=%d dt=%s written=%d)",
			req.Method,
			req.URL,
			m.Code,
			m.Duration,
			m.Written,
		)
	})
}

// ListenAndServe start listing http server
func ListenAndServe(router *mux.Router) {
	log.Printf("Web: Preping ")

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", system.Get("http_address", "127.0.0.1"), system.Get("http_port", "80")),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Web: Started server on %s and listening ... ", fmt.Sprintf("%s:%s", system.Get("http_address", ""), system.Get("http_port", "80")))
	log.Fatal(s.ListenAndServe())
}
