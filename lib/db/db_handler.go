package db

import (
	"database/sql"
	"fmt"
	"log"

	"upsilon_cities_go/config"

	// needed for postgres driver
	"github.com/lib/pq"
)

// Handler Contains DB related informations
type Handler struct {
	db   *sql.DB
	open bool
}

// New Create a new handler for database, ensure database is created
func New() *Handler {
	handler := new(Handler)
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		config.DB_USER, config.DB_PASSWORD, config.DB_NAME, config.DB_HOST, config.DB_PORT)

	db, _ := sql.Open("postgres", dbinfo)

	errPing := db.Ping()
	if err, ok := errPing.(*pq.Error); ok {
		log.Fatalf("DB: Database failed to be connected: %s", err)
	} else {
		log.Printf("DB: Successfully connected to : %s %s", config.DB_HOST, config.DB_NAME)
	}

	handler.db = db
	handler.open = true
	return handler
}

// Exec executes provided query and check if it's correctly executed or not.
// Abort app if not.
func (dbh *Handler) Exec(query string) (result *sql.Rows) {
	dbh.CheckState()
	log.Printf("DB: About to Exec: %s", query)
	result, err := dbh.db.Query(query)
	errorCheck(query, err)
	return result
}

// Query Just like Exec but uses Postgres formater.
func (dbh *Handler) Query(format string, a ...interface{}) (result *sql.Rows) {
	dbh.CheckState()
	log.Printf("DB: About to Query: %s", format)
	result, err := dbh.db.Query(format, a...)
	errorCheck(format, err)
	return result
}

// CheckState assert that connection to DB is still alive. or break
func (dbh *Handler) CheckState() {
	if !dbh.open {
		log.Fatal("DB: Can't use this connection, it's been closed")
	}
	err := dbh.db.Ping()
	if err != nil {
		log.Fatalf("DB: Can't use this connection, an error occured: %s", err)
	}
}

// Close frees db ressource
func (dbh *Handler) Close() {
	if dbh.open {
		dbh.open = false
		defer dbh.db.Close()
	} else {
		log.Print("DB: Already Closed")
	}
}

// ErrorCheck checks if query result has an error or not
func errorCheck(query string, err error) bool {
	if err != nil {
		log.Printf("DB: Failed to execute query: %s", query)

		// fatal aborts app
		log.Fatalf("DB: Aborting: %s", err)

		return true
	}

	return false
}
