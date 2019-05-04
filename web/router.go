package web

import (
	"log"
	"net/http"
	"time"
	"upsilon_cities_go/config"
	grid_controller "upsilon_cities_go/web/controllers/grid"
	"upsilon_cities_go/web/templates"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
)

// RouterSetup Prepare routing.
func RouterSetup() *mux.Router {
	r := mux.NewRouter()

	// CRUD /maps
	r.HandleFunc("/map/{map_id}", grid_controller.Show).Methods("GET")
	r.HandleFunc("/map", grid_controller.Index).Methods("GET")
	r.HandleFunc("/map", grid_controller.Create).Methods("POST")

	// JSON Access ...
	jsonAPI := r.PathPrefix("/api").Subrouter()
	jsonAPI.HandleFunc("/map/{map_id}", grid_controller.Show).Methods("GET")
	jsonAPI.HandleFunc("/map", grid_controller.Index).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.STATIC_FILES))))

	r.Use(logResultMw)
	r.Use(loggingMw)
	return r
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
		Addr:           ":80",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Web: Started server on 127.0.0.1:80 and listening ... ")
	s.ListenAndServe()
}
