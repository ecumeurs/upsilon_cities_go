package web

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"upsilon_cities_go/config"
	"upsilon_cities_go/lib/db"
	controllers "upsilon_cities_go/web/controllers"
	crv_controller "upsilon_cities_go/web/controllers/caravan"
	city_controller "upsilon_cities_go/web/controllers/city"
	corp_controller "upsilon_cities_go/web/controllers/corporation"
	grid_controller "upsilon_cities_go/web/controllers/grid"
	user_controller "upsilon_cities_go/web/controllers/user"

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
	store, err = pgstore.NewPGStoreFromPool(dbh.Raw(), []byte(config.SESSION_SECRET_KEY))
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

	city := sessionned.PathPrefix("/city/{city_id}").Subrouter()
	city.HandleFunc("", city_controller.Show).Methods("GET")
	city.HandleFunc("/producer/{producer_id}/{action}", city_controller.ProducerUpgrade).Methods("POST")

	corporation := sessionned.PathPrefix("/corporation/{corp_id}").Subrouter()
	corporation.HandleFunc("/", corp_controller.Show).Methods("GET")

	caravan := sessionned.PathPrefix("/caravan").Subrouter()
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
	caravan.HandleFunc("/{crv_id}/drop", crv_controller.Abort).Methods("POST")

	usr := sessionned.PathPrefix("/user").Subrouter()
	usr.HandleFunc("", user_controller.Show).Methods("GET")
	usr.HandleFunc("/new", user_controller.New).Methods("GET")
	usr.HandleFunc("", user_controller.Create).Methods("POST")
	usr.HandleFunc("/login", user_controller.ShowLogin).Methods("GET")
	usr.HandleFunc("/login", user_controller.Login).Methods("POST")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("GET")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("POST")
	usr.HandleFunc("/reset_password", user_controller.ShowResetPassword).Methods("GET")
	usr.HandleFunc("/reset_password", user_controller.ResetPassword).Methods("POST")
	usr.HandleFunc("", user_controller.Destroy).Methods("DELETE")

	admin := sessionned.PathPrefix("/admin").Subrouter()
	adminUser := admin.PathPrefix("/users").Subrouter()
	adminUser.HandleFunc("", user_controller.Index).Methods("GET")
	adminUser.HandleFunc("/{user_id}", user_controller.AdminShow).Methods("GET")
	adminUser.HandleFunc("/{user_id}/reset", user_controller.AdminReset).Methods("POST")
	adminUser.HandleFunc("/{user_id}", user_controller.AdminDestroy).Methods("DELETE")

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

	city = jsonAPI.PathPrefix("/city/{city_id}").Subrouter()
	city.HandleFunc("", city_controller.Show).Methods("GET")
	city.HandleFunc("/producer/{producer_id}/{action}/{product}", city_controller.ProducerUpgrade).Methods("POST")

	usr = jsonAPI.PathPrefix("/user").Subrouter()
	usr.HandleFunc("", user_controller.Show).Methods("GET")
	usr.HandleFunc("/new", user_controller.New).Methods("GET")
	usr.HandleFunc("", user_controller.Create).Methods("POST")
	usr.HandleFunc("/login", user_controller.ShowLogin).Methods("GET")
	usr.HandleFunc("/login", user_controller.Login).Methods("POST")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("GET")
	usr.HandleFunc("/logout", user_controller.Logout).Methods("POST")
	usr.HandleFunc("/reset_password", user_controller.ShowResetPassword).Methods("GET")
	usr.HandleFunc("/reset_password", user_controller.ResetPassword).Methods("POST")
	usr.HandleFunc("", user_controller.Destroy).Methods("DELETE")

	corporation = jsonAPI.PathPrefix("/corporation/{corp_id}").Subrouter()
	corporation.HandleFunc("/", corp_controller.Show).Methods("GET")

	caravan = jsonAPI.PathPrefix("/caravan").Subrouter()
	// caravan related stuff
	caravan.HandleFunc("", crv_controller.Index).Methods("GET")
	caravan.HandleFunc("/new/{city_id}", crv_controller.New).Methods("GET")
	caravan.HandleFunc("/seek", crv_controller.Seek).Methods("GET")
	caravan.HandleFunc("", crv_controller.Create).Methods("POST")
	caravan.HandleFunc("/{crv_id}", crv_controller.Show).Methods("GET")
	caravan.HandleFunc("/{crv_id}/accept", crv_controller.Accept).Methods("POST")
	caravan.HandleFunc("/{crv_id}/reject", crv_controller.Reject).Methods("POST")
	caravan.HandleFunc("/{crv_id}/abort", crv_controller.Abort).Methods("POST")
	caravan.HandleFunc("/{crv_id}/counter", crv_controller.GetCounter).Methods("POST")
	caravan.HandleFunc("/{crv_id}/counter", crv_controller.PostCounter).Methods("POST")
	caravan.HandleFunc("/{crv_id}/drop", crv_controller.Abort).Methods("POST")

	admin = jsonAPI.PathPrefix("/admin").Subrouter()
	adminUser = admin.PathPrefix("/users").Subrouter()
	adminUser.HandleFunc("", user_controller.Index).Methods("GET")
	adminUser.HandleFunc("/{user_id}", user_controller.AdminShow).Methods("GET")
	adminUser.HandleFunc("/{user_id}/reset", user_controller.AdminReset).Methods("POST")
	adminUser.HandleFunc("/{user_id}", user_controller.AdminDestroy).Methods("DELETE")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.MakePath(config.STATIC_FILES)))))

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.FromSlash(fmt.Sprintf("%s/img/favicon.ico", config.MakePath(config.STATIC_FILES))))
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

		if err = session.Save(r, w); err != nil {
			log.Fatalf("Error saving session: %v", err)
		}
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
		Addr:           config.HTTP_PORT,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Web: Started server on 127.0.0.1%s and listening ... ", config.HTTP_PORT)
	s.ListenAndServe()
}
