package web

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"upsilon_cities_go/config"
	"upsilon_cities_go/lib/db"
	city_controller "upsilon_cities_go/web/controllers/city"
	grid_controller "upsilon_cities_go/web/controllers/grid"
	"upsilon_cities_go/web/templates"

	"github.com/antonlindstrom/pgstore"
	"github.com/felixge/httpsnoop"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// RouterSetup Prepare routing.
func RouterSetup() *mux.Router {
	r := mux.NewRouter()

	dbh := db.New()
	store, err := pgstore.NewPGStoreFromPool(dbh.Raw(), []byte(config.SESSION_SECRET_KEY))
	if err != nil {
		// failed to find a store in there ...
		log.Fatalf("Session: Failed to initialize session for web request ... %s", err)
	}

	defer store.Close()
	// Run a background goroutine to clean up expired sessions from the database.
	defer store.StopCleanup(store.Cleanup(time.Minute * 5))

	// ensure session knows how to keep some complex data types.
	initConverters()

	// CRUD /maps
	r.HandleFunc("/map", grid_controller.Index).Methods("GET")
	r.HandleFunc("/map", grid_controller.Create).Methods("POST")

	maps := r.PathPrefix("/map/{map_id}").Subrouter()
	maps.HandleFunc("", grid_controller.Show).Methods("GET")
	maps.HandleFunc("/cities", city_controller.Index).Methods("GET")
	maps.HandleFunc("/city/{city_id}", city_controller.Show).Methods("GET")

	// JSON Access ...
	jsonAPI := r.PathPrefix("/api").Subrouter()
	jsonAPI.HandleFunc("/map", grid_controller.Index).Methods("GET")
	jsonAPI.HandleFunc("/map", grid_controller.Create).Methods("POST")

	maps = jsonAPI.PathPrefix("/map/{map_id}").Subrouter()
	maps.HandleFunc("", grid_controller.Show).Methods("GET")
	maps.HandleFunc("", grid_controller.Destroy).Methods("DELETE")
	maps.HandleFunc("/cities", city_controller.Index).Methods("GET")
	maps.HandleFunc("/city/{city_id}", city_controller.Show).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.MakePath(config.STATIC_FILES)))))

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Web: Prepring favicon")
		http.ServeFile(w, r, filepath.FromSlash(fmt.Sprintf("%s/img/favicon.ico", config.MakePath(config.STATIC_FILES))))
	})

	r.Use(logResultMw)
	r.Use(loggingMw)
	r.Use(sessionMw)
	return r
}

// initialize "gob" that handle struct serialization for session.
// see: https://www.gorillatoolkit.org/pkg/sessions#overview
func initConverters() {

}

//sessionMw start a session
func sessionMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		dbh := db.New()
		store, err := pgstore.NewPGStoreFromPool(dbh.Raw(), []byte(config.SESSION_SECRET_KEY))
		if err != nil {
			// failed to find a store in there ...
			log.Fatalf("Session: Failed to initialize session for web request ... %s", err)
		}

		// closes also dbh
		defer store.Close()
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
	templates.LoadTemplates()

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
